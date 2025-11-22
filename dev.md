goctl api go -api product/api/product.api -dir product/api/
goctl api plugin -plugin goctl-swagger="swagger -filename product.json -host localhost:8888 -basepath /" -api product.api -dir .

goctl model mysql datasource `
-url="root:darling1224@tcp(127.0.0.1:3306)/wmss" `
-table="sys_*" `
-dir="./model" `
-cache=false \ `
--style=goZero `

goctl api go -api user.api -dir ./

goctl api plugin -plugin goctl-swagger="swagger -filename user.json -host localhost:8889 -basepath /" -api user.api -dir .

admin:123456