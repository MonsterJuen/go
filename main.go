package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/caict-4iot-dev/BIF-Core-SDK-Go/module/contract"
	"github.com/caict-4iot-dev/BIF-Core-SDK-Go/types/request"
	"github.com/gin-gonic/gin"
)

type Product struct {
	ProductID       string `json:"productId"`
	ProductName     string `json:"productName"`
	Origin          string `json:"origin"`
	ProcessLocation string `json:"processLocation"`
	ProcessTime     string `json:"processTime"`
	Value           string `json:"value"`
	StorageTime     string `json:"storageTime"`
}

// 调用合约input数据体
type ContractCall struct {
	Function string `json:"function"`
	Args     string `json:"args"`
}

// 查询合约信息input数据体
type ContractCa struct {
	Function string `json:"function"`
	Args     string `json:"args"`
	Return   string `json:"return"`
}

// 解析返回体
type ContractResponse struct {
	ErrorCode int    `json:"error_code"`
	ErrorDesc string `json:"error_desc"`
	Result    struct {
		QueryRets []struct {
			Result struct {
				Data string `json:"data"` // 目标字段
			} `json:"result"`
		} `json:"query_rets"`
	} `json:"result"`
}

var MyPrivateKey string = "priSPKkpuYHiwQ886GdRrb9s6TbCmTqYdQdKEYo1X6njuSiMNP"
var SDK_URL string = "http://test.bifcore.bitfactory.cn"
var MyAccountAddress string = "did:bid:efPLdVAy6AN5wVgViFzfeNZ5yauq7hFs"
var ContractAddress = "did:bid:efZAgTKiwi84PoQkmkaSkP9w54yBL1N8"

func GenerateCall(s Product) ContractCall {
	return ContractCall{
		Function: "createProduct(uint256,string,string,string,string,uint256,string)",
		Args: fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s",
			s.ProductID,
			s.ProductName,
			s.Origin,
			s.ProcessLocation,
			s.ProcessTime,
			s.Value,
			s.StorageTime,
		),
	}
}

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

func Querycontrct(senderAddress string, contractAddress string, id string) (*Product, error) {
	// 1. 初始化合约实例（连接区块链节点）
	bs := contract.GetContractInstance(SDK_URL) // SDK_INSTANCE_URL 需替换为实际节点地址

	call := ContractCa{
		Function: "getProduct(uint256)",
		// Args:     fmt.Sprintf("%s", id),
		Args:   id,
		Return: "returns(uint256,string,string,string,string,uint256,string)",
	}
	// 序列化为JSON字符串
	jsonBytes, err := json.Marshal(call)
	if err != nil {
		panic(err)
	}

	// 2. 构建合约调用请求参数
	var r request.BIFContractCallRequest

	// 4. 填充请求参数
	r.SourceAddress = senderAddress     // 指定交易发送者
	r.ContractAddress = contractAddress // 要调用的合约地址
	r.FeeLimit = 100000000
	r.Input = string(jsonBytes)
	r.Type = 1

	// 5. 调用合约方法
	res := bs.ContractQuery(r) // 发送交易到区块链网络
	if res.ErrorCode != 0 {
		// 错误处理：输出错误详情并终止程序
		log.Fatalf("合约调用失败 (错误码=%d): %s", res.ErrorCode, res.ErrorDesc)
	}

	// // 6. 序列化结果并打印
	dataByte, err := json.Marshal(res)
	if err != nil {
		log.Fatalf("JSON序列化失败: %v", err)
	}

	var response ContractResponse

	// 反序列化到结构体
	if err := json.Unmarshal(dataByte, &response); err != nil {
		log.Fatalf("JSON解析失败: %v", err)
	}

	// 提取data字段
	if len(response.Result.QueryRets) == 0 {
		log.Fatal("无查询结果")
	}
	rawData := response.Result.QueryRets[0].Result.Data
	fmt.Println("原始data字符串:", rawData)
	product, err := parseProductData(rawData)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func parseProductData(rawData string) (*Product, error) {
	// 去除首尾的方括号
	trimmed := strings.Trim(rawData, "[]")
	// 按逗号分割字符串
	parts := strings.Split(trimmed, ",")
	// 检查字段数量
	if len(parts) < 7 {
		return nil, fmt.Errorf("数据字段不足，期望7，实际%d", len(parts))
	}
	// 去除每个字段的前后空格
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	// 构造结构体
	product := &Product{
		ProductID:       parts[0],
		ProductName:     parts[1],
		Origin:          parts[2],
		ProcessLocation: parts[3],
		ProcessTime:     parts[4],
		Value:           parts[5],
		StorageTime:     parts[6],
	}
	return product, nil
}

func main() {
	r := gin.Default()

	// 设置静态文件目录
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 路由设置
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/query", func(c *gin.Context) {
		c.HTML(http.StatusOK, "query.html", nil)
	})

	r.POST("/api/product", func(c *gin.Context) {
		var product Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//
		_, err := Querycontrct(MyAccountAddress, ContractAddress, product.ProductID)

		if err != nil {

			// 上链操作
			contractCall := GenerateCall(product)
			jsonBytes, _ := json.Marshal(contractCall)
			input := string(jsonBytes)
			fmt.Println(input)
			ContractCalls(MyAccountAddress, ContractAddress, MyPrivateKey, input)
			c.JSON(http.StatusOK, gin.H{
				"message": "产品信息保存成功",
				"data":    product,
			})

			return

		} else {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "该产品id已存在"})
			return

		}
		// // 上链操作
		// contractCall := GenerateCall(product)
		// jsonBytes, _ := json.Marshal(contractCall)
		// input := string(jsonBytes)
		// fmt.Println(input)
		// createrr := ContractCalls(MyAccountAddress, ContractAddress, MyPrivateKey, input)
		// if createrr != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "该产品id已存在"})
		// 	return
		// }

		// c.JSON(http.StatusOK, gin.H{
		// 	"message": "产品信息保存成功",
		// 	"data":    product,
		// })

	})

	// 添加查询路由
	r.GET("/api/product/:id", func(c *gin.Context) {
		productId := c.Param("id")
		//查询操作
		product, err := Querycontrct(MyAccountAddress, ContractAddress, productId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到该水产品信息"})
			return
		}

		c.JSON(http.StatusOK, product)
	})

	r.Run(":8090")
}
