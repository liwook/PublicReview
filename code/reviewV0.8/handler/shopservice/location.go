package shopservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"review/dal/model"
	"review/dal/query"
	"review/db"
	"review/pkg/response"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	geoKeyPrefix         = "geo:"
	defaultStoresPerPage = 10

	queryLongitude   = "longitude"
	queryLatitude    = "latitude"
	queryDistance    = "distance"
	queryTypeId      = "typeId"
	queryCurrentPage = "currentPage"
	rawShopSql       = "select * from tb_shop where type_id = ? and id in (%s) order by field(id,%s)"
)

const (
	maxLongitude = 180.0
	minLongitude = -180.0
	maxLatitude  = 90.0
	minLatitude  = -90.0
	maxDistance  = 100.0 // 最大查询距离（公里）
)

func LoadShopListToCache() error {
	tbshop := query.TbShop
	shops, err := tbshop.Select(tbshop.ID, tbshop.X, tbshop.Y, tbshop.TypeID).Find()
	if err != nil {
		slog.Error("LoadShopListToCache error", "err", err)
		return err
	}
	//将shop按照typeId进行分类
	shopMap := make(map[uint64][]*model.TbShop)
	for _, shop := range shops {
		shopMap[shop.TypeID] = append(shopMap[shop.TypeID], shop)
	}

	//使用管道，一次性输入
	pipeline := db.RedisDb.Pipeline()
	for _, shops := range shopMap {
		for _, shop := range shops {
			pipeline.GeoAdd(context.Background(), geoKeyPrefix+strconv.Itoa(int(shop.TypeID)), &redis.GeoLocation{Longitude: shop.X, Latitude: shop.Y, Name: strconv.FormatUint(shop.ID, 10)})
		}
	}
	_, err = pipeline.Exec(context.Background())
	if err != nil {
		slog.Error("Failed to execute Redis pipeline", "err", err)
		return err
	}
	return nil
}

// GET /api/v1/shop/distance-list
func QueryShopDistance(c *gin.Context) {
	params, err := parseAndValidateParams(c)
	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}

	var result []shopLoationDTO
	if params.UseGeoQuery {
		result, err = queryShopByGeoLocation(params)
	} else {
		result, err = queryShopByType(params)
	}

	response.HandleBusinessResult(c, err, result)
}

func parseAndValidateParams(c *gin.Context) (*shopQueryParams, error) {
	params := &shopQueryParams{}

	// 验证必需参数
	typeIdStr := c.Query(queryTypeId)
	if typeIdStr == "" {
		return nil, response.NewBusinessError(response.ErrValidation, "typeId 参数不能为空")
	}

	id, err := strconv.Atoi(typeIdStr)
	if err != nil || id <= 0 {
		return nil, response.NewBusinessError(response.ErrValidation, "typeId 必须为正整数")
	}
	params.TypeID = id

	// 验证分页参数
	currentPageStr := c.Query(queryCurrentPage)
	if currentPageStr == "" {
		params.CurrentPage = 1 // 默认第一页
	} else {
		page, err := strconv.Atoi(currentPageStr)
		if err != nil || page < 1 {
			return nil, response.NewBusinessError(response.ErrValidation, "currentPage 必须为大于等于 1 的整数")
		}
		params.CurrentPage = page
	}

	// 检查是否需要地理位置查询
	longitude := c.Query(queryLongitude)
	latitude := c.Query(queryLatitude)
	distance := c.Query(queryDistance)

	// 如果任一地理位置参数为空，则使用普通查询
	if longitude == "" || latitude == "" || distance == "" {
		params.UseGeoQuery = false
		return params, nil
	}

	// 验证地理位置参数
	if err := validateGeoParams(longitude, latitude, distance, params); err != nil {
		return nil, err
	}

	params.UseGeoQuery = true
	return params, nil
}

func validateGeoParams(longitude, latitude, distance string, params *shopQueryParams) error {
	// 验证距离
	distanceFloat, err := strconv.ParseFloat(distance, 64)
	if err != nil || distanceFloat <= 0 || distanceFloat > maxDistance {
		return response.NewBusinessError(response.ErrValidation, fmt.Sprintf("distance 必须为 0 到 %.1f 之间的正数", maxDistance))
	}
	params.Distance = distanceFloat

	// 验证经度
	lng, err := strconv.ParseFloat(longitude, 64)
	if err != nil || lng < minLongitude || lng > maxLongitude {
		return response.NewBusinessError(response.ErrValidation, fmt.Sprintf("longitude 必须在 %.1f 到 %.1f 之间", minLongitude, maxLongitude))
	}
	params.Longitude = lng

	// 验证纬度
	lat, err := strconv.ParseFloat(latitude, 64)
	if err != nil || lat < minLatitude || lat > maxLatitude {
		return response.NewBusinessError(response.ErrValidation, fmt.Sprintf("latitude 必须在 %.1f 到 %.1f 之间", minLatitude, maxLatitude))
	}
	params.Latitude = lat

	return nil
}

func queryShopByType(params *shopQueryParams) ([]shopLoationDTO, error) {
	tbShop := query.TbShop
	offset := (params.CurrentPage - 1) * defaultStoresPerPage
	shops, err := tbShop.Where(tbShop.TypeID.Eq(uint64(params.TypeID))).Limit(defaultStoresPerPage).Offset(offset).Find()
	if err != nil {
		return nil, response.WrapBusinessError(response.ErrDatabase, err, "查询店铺失败")
	}

	return convertTbShopPtrToResponse(shops), nil
}

func queryShopByGeoLocation(params *shopQueryParams) ([]shopLoationDTO, error) {
	from := (params.CurrentPage - 1) * defaultStoresPerPage
	end := params.CurrentPage * defaultStoresPerPage
	res, err := db.RedisDb.GeoSearchLocation(context.Background(), geoKeyPrefix+strconv.Itoa(params.TypeID), &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  params.Longitude,
			Latitude:   params.Latitude,
			Radius:     params.Distance,
			RadiusUnit: "km",
			Sort:       "ASC", //升序，从小排到大
			Count:      end,
		},

		WithDist: true, //返回距离，用于显示
		// WithCoord: true, //返回经纬度， 数据库中有，所以不需要返回了
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, response.NewBusinessError(response.ErrNotFound, "没有找到相关店铺")
		}
		return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
	}

	if len(res) <= from { //表示没有下一页了，直接返回空
		return []shopLoationDTO{}, nil
	}
	if len(res) < end {
		end = len(res)
	}
	tmp := res[from:end] //截取需要的数据

	distanceMap := make(map[uint64]float64) //key是shopId，value是距离
	shopIds := make([]uint64, 0, len(tmp))
	for _, v := range tmp {
		id, err := strconv.ParseUint(v.Name, 10, 64)
		if err != nil {
			slog.Error("ParseUint error", "err", err, "name", v.Name)
			continue
		}
		shopIds = append(shopIds, id)
		distanceMap[id] = v.Dist
	}

	// 构建安全的参数化查询,防止sql注入
	placeholders := make([]string, len(shopIds))
	args := make([]any, len(shopIds)*2)
	for i, id := range shopIds {
		placeholders[i] = "?"
		args[i] = id              // 用于IN子句
		args[len(shopIds)+i] = id // 用于FIELD函数
	}
	inClause := strings.Join(placeholders, ",")

	var shops []model.TbShop
	err = db.DBEngine.Raw(fmt.Sprintf(rawShopSql, inClause, inClause), params.TypeID, args).Scan(&shops).Error
	if err != nil {
		slog.Error(err.Error())
		return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
	}

	result := convertTbShopToResponse(shops)
	for i, shop := range result {
		if dist, exists := distanceMap[shop.ID]; exists {
			result[i].Distance = dist
		}
	}

	return result, nil
}
