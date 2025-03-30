package models

import "time"

// Goods 货物模型
type Goods struct {
	ID            string    `json:"id"`
	GoodsID       string    `json:"goodsId"`       // 批次编号
	Name          string    `json:"name"`          // 海鲜名称
	Type          string    `json:"type"`          // 类型
	Specification string    `json:"specification"` // 规格
	Weight        float64   `json:"weight"`        // 重量(吨)
	Temperature   float64   `json:"temperature"`   // 温度(°C)
	Location      string    `json:"location"`      // 存储位置
	InTime        time.Time `json:"inTime"`        // 入库时间
	Status        string    `json:"status"`        // 状态
}

// GoodsFilter 货物查询过滤器
type GoodsFilter struct {
	GoodsID string `json:"goodsId"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

// GoodsResponse 货物列表响应
type GoodsResponse struct {
	Total int64   `json:"total"`
	Items []Goods `json:"items"`
}
