package client

// import (
// 	"io/ioutil"
// 	"net/http"
// 	"strconv"
// 	"strings"
//
// 	"github.com/Appscrunch/Multy-back/currencies"
// 	"github.com/Appscrunch/Multy-back/store"
// 	"github.com/gin-gonic/gin"
//
// 	"gopkg.in/mgo.v2/bson"
// )
//
// type Total struct {
// 	Assets map[string]string `json:"displayWallet"` // name of wallet to ballance
// }
//
// func getaddresses(c *gin.Context, token string) map[int]DisplayWallet {
// 	sel := bson.M{"devices.JWT": token}
//
// 	user := store.User{}
//
// 	usersData.Find(sel).One(&user)
//
// 	userWallets := map[int]DisplayWallet{}
// 	if user.UserID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"code":    400,
// 			"message": http.StatusText(400),
// 		})
// 		return userWallets
// 	}
//
// 	for k, v := range user.Wallets {
// 		userWallets[k] = DisplayWallet{v.Chain, v.Adresses}
// 	}
//
// 	return userWallets
// }
//
// func getexchangeprice(c *gin.Context, from, to string) float64 {
// 	url := "https://min-api.cryptocompare.com/data/price?fsym=" + strings.ToUpper(from) + "&tsyms=" + strings.ToUpper(to)
// 	var er map[string]interface{}
// 	makeRequest(c, url, &er)
// 	f, _ := er[to].(float64)
// 	return f
// }
//
// func getaddressbalance(c *gin.Context, addr string, chain int) int {
// 	urls := []string{
// 		"https://blockchain.info/q/addressbalance/" + addr,                   //bitcoin
// 		"https://api.blockcypher.com/v1/eth/main/addrs/" + addr + "/balance", //etherium
// 	}
// 	var to map[string]interface{}
//
// 	switch chain {
// 	case currencies.Bitcoin:
// 		response, err := http.Get(urls[0])
// 		responseErr(c, err, http.StatusServiceUnavailable) // 503
//
// 		data, err := ioutil.ReadAll(response.Body)
// 		responseErr(c, err, http.StatusInternalServerError) // 500
//
// 		b, _ := strconv.Atoi(string(data))
// 		return b
//
// 	case currencies.Ether:
// 		makeRequest(c, urls[1], &to)
// 		balance := to["balance"]
// 		i := balance.(int)
// 		return i
// 	default:
// 		return 0
// 	}
//
// }
