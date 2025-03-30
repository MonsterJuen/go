package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"storageSys/models"
	"time"

	"github.com/caict-4iot-dev/BIF-Core-SDK-Go/module/contract"
	"github.com/caict-4iot-dev/BIF-Core-SDK-Go/types/request"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 调用合约input数据体
type ContractCall struct {
	Function string `json:"function"`
	Args     string `json:"args"`
}

var MyPrivateKey string = "priSPKkpuYHiwQ886GdRrb9s6TbCmTqYdQdKEYo1X6njuSiMNP"
var SDK_URL string = "http://test.bifcore.bitfactory.cn"
var MyAccountAddress string = "did:bid:efPLdVAy6AN5wVgViFzfeNZ5yauq7hFs"
var ContractAddress = "did:bid:efwjBFEAAXzdhnhuP8XDBfCVGSJEh2kn"

// 模拟数据存储
var goodsList = make([]models.Goods, 0)

func ContractCalls(senderAddress string, contractAddress string, senderPrivateKey string, input string) {
	// 1. 初始化合约实例（连接区块链节点）
	bs := contract.GetContractInstance(SDK_URL) // SDK_INSTANCE_URL 需替换为实际节点地址

	// 2. 构建合约调用请求参数
	var r request.BIFContractInvokeRequest

	// 4. 填充请求参数
	r.SenderAddress = senderAddress     // 指定交易发送者
	r.PrivateKey = senderPrivateKey     // 私钥用于签名交易
	r.ContractAddress = contractAddress // 要调用的合约地址
	r.BIFAmount = 0                     // 转账金额（单位：链的最小单位，如 1 BIF = 1e8 最小单位）
	r.Input = input
	r.Remarks = "contract invoke" // 交易备注（可读描述）
	r.FeeLimit = 100000000
	res := bs.ContractInvoke(r)
	fmt.Println(res.ErrorCode)

}

func GenerateCall(s models.Goods) ContractCall {
	return ContractCall{
		Function: "createProduct(string,string,string,string,string,string,string,string,string)",
		Args: fmt.Sprintf("%s,%s,%s,%s,%f,%f,%s,%s,%s",
			s.GoodsID,
			s.Name,
			s.Type,
			s.Specification,
			s.Weight,
			s.Temperature,
			s.Location,
			s.InTime,
			s.Status,
		),
	}
}

// GetGoodsList 获取货物列表
func GetGoodsList(c *gin.Context) {
	var filter models.GoodsFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 过滤数据
	filteredList := make([]models.Goods, 0)
	for _, goods := range goodsList {
		if filter.GoodsID != "" && goods.GoodsID != filter.GoodsID {
			continue
		}
		if filter.Type != "" && goods.Type != filter.Type {
			continue
		}
		if filter.Status != "" && goods.Status != filter.Status {
			continue
		}
		filteredList = append(filteredList, goods)
	}

	// 分页
	start := (filter.Page - 1) * filter.Size
	end := start + filter.Size
	if end > len(filteredList) {
		end = len(filteredList)
	}
	if start > len(filteredList) {
		start = len(filteredList)
	}

	response := models.GoodsResponse{
		Total: int64(len(filteredList)),
		Items: filteredList[start:end],
	}

	c.JSON(http.StatusOK, response)
}

// CreateInbound 创建入库记录
func CreateInbound(c *gin.Context) {
	var goods models.Goods
	if err := c.ShouldBindJSON(&goods); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一ID和批次编号
	goods.ID = uuid.New().String()
	goods.GoodsID = "SF" + time.Now().Format("20060102") + uuid.New().String()[:3]
	goods.InTime = time.Now()
	goods.Status = "in_stock"

	contractCall := GenerateCall(goods)
	jsonBytes, _ := json.Marshal(contractCall)
	input := string(jsonBytes)
	fmt.Println(input)
	ContractCalls(MyAccountAddress, ContractAddress, MyPrivateKey, input)

	c.JSON(http.StatusOK, goods)
}

// UpdateGoods 更新货物信息
func UpdateGoods(c *gin.Context) {
	id := c.Param("id")
	var updatedGoods models.Goods
	if err := c.ShouldBindJSON(&updatedGoods); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, goods := range goodsList {
		if goods.ID == id {
			updatedGoods.ID = id
			updatedGoods.GoodsID = goods.GoodsID
			updatedGoods.InTime = goods.InTime
			goodsList[i] = updatedGoods
			c.JSON(http.StatusOK, updatedGoods)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "goods not found"})
}

// DeleteGoods 删除货物
func DeleteGoods(c *gin.Context) {
	id := c.Param("id")

	for i, goods := range goodsList {
		if goods.ID == id {
			goodsList = append(goodsList[:i], goodsList[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "goods not found"})
}

// OutboundGoods 货物出库
func OutboundGoods(c *gin.Context) {
	id := c.Param("id")

	for i, goods := range goodsList {
		if goods.ID == id {
			if goods.Status != "in_stock" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "goods is not in stock"})
				return
			}
			goodsList[i].Status = "out_stock"
			c.JSON(http.StatusOK, goodsList[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "goods not found"})
}

// MortgageGoods 货物抵押
func MortgageGoods(c *gin.Context) {
	id := c.Param("id")

	for i, goods := range goodsList {
		if goods.ID == id {
			if goods.Status != "in_stock" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "goods is not in stock"})
				return
			}
			goodsList[i].Status = "mortgaged"
			c.JSON(http.StatusOK, goodsList[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "goods not found"})
}
