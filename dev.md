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

goctl model mysql datasource -url="root:yuchao123@tcp(localhost:3306)/wmss" -table="product_net_value,purchase_application,redemption_application,customer_position,transaction_confirmation,liquidation_log,work_calendar" -dir ./model -cache=false -style go_zero

goctl model mysql datasource -url="root:yuchao123@tcp(127.0.0.1:3306)/wmss" -table="purchase_application,redemption_application,transaction_confirmation,customer_position,product_info,customer_info,customer_bank_card,product_net_value" -dir ./model cache=false -style go_zero