package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/smallnest/rpcx/share"
	"go.uber.org/zap"
	"ser/db"
	"strings"
	"time"
)

type Customer struct {
	logger *zap.Logger
}

type UserInfo struct {
	UserId   int                    `json:"userId"`   //用户ID
	UserType int8                   `json:"userType"` //1后台用户  2前端客户
	Data     map[string]interface{} `json:"data"`
}

type Result struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

//根据元数据解析获取用户信息
func GetUserInfo(ctx context.Context) UserInfo {
	reqMeta := ctx.Value(share.ReqMetaDataKey).(map[string]string)
	var userInfo UserInfo
	json.Unmarshal([]byte(reqMeta["userInfo"]), &userInfo)
	return userInfo
}
//新增地址接收请求参数
type CreateCustomerAddressReq struct {
	CustomerId int    `json:"customerId,omitempty"`
	Name       string `json:"name" validate:"required" label:"收货人名字"`             //收货人名字
	Phone      string `json:"phone" validate:"required,checkPhone" label:"收货人电话"` //收货人电话
	FixedTel   string `json:"fixedTel" label:"固定电话"`                              //固定电话
	ProvinceId int `json:"provinceId" validate:"required" label:"省份"`          //省份id
	CityId     int `json:"cityId" validate:"required" label:"城市"`              //城市id
	AreaId     int `json:"areaId" validate:"required" label:"区域"`              //区域id
	Address    string `json:"address" validate:"required" label:"详细地址"`           //详细地址
	IsDefault  int8 `json:"isDefault,omitempty"  label:"是否默认"`                  //是否默认
}
//数据库实体
type CustomerAddress struct {
	Id         int `gorm:"primary_key"`
	CustomerId int
	Name       string         //收货人名字
	Phone      string         //收货人电话
	FixedTel   sql.NullString //固定电话
	ProvinceId int         //省份id
	CityId     int         //城市id
	AreaId     int         //区域id
	Address    string         //详细地址
	IsDefault  int8         //是否默认 1:否 2:是
	CreateTime time.Time //创建时间
}

func CheckNullStr(str string) sql.NullString {
	str = strings.Trim(str, " ")
	if str == "" {
		return sql.NullString{
			String: "",
			Valid:  false,
		}
	} else {
		return sql.NullString{
			String: str,
			Valid:  true,
		}
	}
}



//新增Customer 收货地址 网站/管理系统 通用
func (c *Customer) CreateCustomerAddress(ctx context.Context, req *CreateCustomerAddressReq, rsp *Result) error {
	userInfo := GetUserInfo(ctx)
	if userInfo.UserType == 2 {
		req.CustomerId = userInfo.UserId
	} else {
		if req.CustomerId == 0 {
			rsp.Success = false
			rsp.Msg = "CustomerId不能为空"
			c.logger.Error("Rule_Error", zap.String("规则错误", "后台调用customer为空"))
			return nil
		}
	}
	var addressCount int64
	var addressNum int64
	err := db.GormDB.Table("customer_address").Where("customer_id = ?", req.CustomerId).Count(&addressNum).Error
	if err != nil {
		rsp.Success = false
		rsp.Msg = err.Error()
		c.logger.Error("Mysql_Error", zap.Error(err))
		return nil
	}
	if addressNum >= 5 {
		rsp.Success = false
		rsp.Msg = "默认最多五个地址"
		c.logger.Error("Rule_Error", zap.String("规则错误", "默认最多五个地址"))
		return nil
	}

	if addressNum == 0 {
		req.IsDefault =2
	}
	err = db.GormDB.Table("province_city t1").Joins("left join province_city t2 on t1.id = t2.parent_id").
		Joins("left join province_city t3 on t2.id = t3.parent_id").
		Where("t1.id = ? and t2.id = ? and t3.id = ?", req.ProvinceId, req.CityId, req.AreaId).Count(&addressCount).Error
	if err != nil {
		rsp.Success = false
		rsp.Msg = err.Error()
		c.logger.Error("Mysql_Error", zap.Error(err))
		return nil
	}
	if addressCount == 0 {
		rsp.Success = false
		rsp.Msg = "所在地区有误,请重新选择"
		c.logger.Error("Rule_Error", zap.String("规则错误", "所在地区有误"))
		return nil
	}
	tx := db.GormDB.Begin()
	if req.IsDefault == 2 {
		err = tx.Table("customer_address").Where("customer_id = ?", req.CustomerId).Update("is_default", 1).Error
		if err != nil {
			tx.Rollback()
			rsp.Success = false
			rsp.Msg = err.Error()
			c.logger.Error("Mysql_Error", zap.Error(err))
			return nil
		}
	}


	var customerAddress = CustomerAddress{
		CustomerId: req.CustomerId,
		Name:       req.Name,
		Phone:      req.Phone,
		CityId:    req.CityId,
		ProvinceId:req.ProvinceId,
		Address:    req.Address,
		FixedTel:   CheckNullStr(req.FixedTel),
		AreaId:    req.AreaId,
		IsDefault:  int8(req.IsDefault),
		CreateTime: time.Now(),
	}
	err = tx.Create(&customerAddress).Error
	if err != nil {
		tx.Rollback()
		rsp.Success = false
		rsp.Msg = err.Error()
		c.logger.Error("Mysql_Error", zap.Error(err))
		return nil
	}
	tx.Commit()
	rsp.Success = true
	rsp.Msg = "新增成功"
	return nil
}
