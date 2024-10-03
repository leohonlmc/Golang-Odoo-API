// package main

// import (
// 	"fmt"
// 	"log"

// 	"github.com/kolo/xmlrpc"
// )

// func getOdoo() {
// 	client, err := xmlrpc.NewClient("https://demo.odoo.com/start", nil)
// 	if err != nil {
// 		log.Fatal("Error creating XML-RPC client:", err)
// 	}

// 	info := map[string]string{}
// 	err = client.Call("start", nil, &info)
// 	if err != nil {
// 		log.Fatal("Error starting Odoo demo instance:", err)
// 	}

// 	url, urlOk := info["host"]
// 	db, dbOk := info["database"]
// 	username, userOk := info["user"]
// 	password, passOk := info["password"]

// 	if !urlOk || !dbOk || !userOk || !passOk {
// 		log.Fatal("Failed to retrieve Odoo instance information")
// 	}

// 	fmt.Println("URL:", url)
// 	fmt.Println("Database:", db)
// 	fmt.Println("Username:", username)
// 	fmt.Println("Password:", password)

// 	loggin(url, db, username, password)
// }

// func loggin(url string, db string, username string, password string) {
// 	var uid int64

// 	client, err := xmlrpc.NewClient(fmt.Sprintf("%s/xmlrpc/2/common", url), nil)
// 	if err != nil {
// 		log.Fatal("Error creating XML-RPC client for authentication:", err)
// 	}

// 	common := map[string]any{}
// 	err = client.Call("version", nil, &common)
// 	if err != nil {
// 		log.Fatal("Error calling 'version' method:", err)
// 	}

// 	fmt.Println("Odoo version information:")
// 	for key, value := range common {
// 		fmt.Printf("%s: %v\n", key, value)
// 	}

// 	err = client.Call("authenticate", []any{
// 		db, username, password,
// 		map[string]any{},
// 	}, &uid)
// 	if err != nil {
// 		log.Fatal("Authentication failed:", err)
// 	}

// 	fmt.Println("Logged in as user ID:", uid)

// 	callMethod(url, db, uid, password)
// }

// func callMethod(url string, db string, uid int64, password string) {
// 	models, err := xmlrpc.NewClient(fmt.Sprintf("%s/xmlrpc/2/object", url), nil)
// 	if err != nil {
// 		log.Fatal("Error creating XML-RPC client for models:", err)
// 	}

// 	var saleOrders []map[string]interface{}
// 	err = models.Call("execute_kw", []interface{}{
// 		db, uid, password,
// 		"sale.order", "search_read",
// 		[][]interface{}{},
// 		map[string]interface{}{"fields": []string{"name", "partner_id", "amount_total"}},
// 	}, &saleOrders)

// 	if err != nil {
// 		log.Fatal("Error fetching sale orders:", err)
// 	}

// 	for _, order := range saleOrders {
// 		fmt.Printf("Sale Order: %s, Customer: %v, Total: %v\n", order["name"], order["partner_id"], order["amount_total"])
// 	}
// }

// func main() {
// 	getOdoo()
// }

// f914d6c6f2a23ae87a880d12619eeaa9a1ff295e

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kolo/xmlrpc"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func getOdoo() {
	loadEnv()

	url := os.Getenv("ODOO_URL")
	db := os.Getenv("ODOO_DB")
	username := os.Getenv("ODOO_USERNAME")
	password := os.Getenv("ODOO_PASSWORD")

	loggin(url, db, username, password)
}

func loggin(url string, db string, username string, password string) {
	var uid int64

	client, err := xmlrpc.NewClient(fmt.Sprintf("%s/xmlrpc/2/common", url), nil)
	if err != nil {
		log.Fatal("Error creating XML-RPC client for authentication:", err)
	}

	err = client.Call("authenticate", []any{
		db, username, password,
		map[string]any{},
	}, &uid)
	if err != nil {
		log.Fatal("Authentication failed:", err)
	}

	fmt.Println("Logged in as user ID:", uid)

	// Call functions to fetch various models
	getSaleOrders(url, db, uid, password)
	getCustomers(url, db, uid, password)
	getProducts(url, db, uid, password)
	getPurchaseOrders(url, db, uid, password)
}

func getSaleOrders(url string, db string, uid int64, password string) {
	fetchAndExport(url, db, uid, password, "sale.order", []string{"name", "partner_id", "amount_total"}, "sale_orders.csv")
}

func getCustomers(url string, db string, uid int64, password string) {
	fetchAndExport(url, db, uid, password, "res.partner", []string{"name", "email", "phone"}, "customers.csv")
}

func getProducts(url string, db string, uid int64, password string) {
	fetchAndExport(url, db, uid, password, "product.product", []string{"name", "list_price", "qty_available"}, "products.csv")
}

func getPurchaseOrders(url string, db string, uid int64, password string) {
	fetchAndExport(url, db, uid, password, "purchase.order", []string{"name", "partner_id", "amount_total"}, "purchase_orders.csv")
}

func fetchAndExport(url, db string, uid int64, password, model string, fields []string, fileName string) {
	models, err := xmlrpc.NewClient(fmt.Sprintf("%s/xmlrpc/2/object", url), nil)
	if err != nil {
		log.Fatal("Error creating XML-RPC client for models:", err)
	}

	var records []map[string]interface{}
	err = models.Call("execute_kw", []interface{}{
		db, uid, password,
		model, "search_read",
		[][]interface{}{},
		map[string]interface{}{
			"fields": fields,
			"limit":  3000,
		},
	}, &records)

	if err != nil {
		log.Fatal(fmt.Sprintf("Error fetching %s records:", model), err)
	}

	exportToCSV(records, fields, fileName)
}

func exportToCSV(records []map[string]interface{}, fields []string, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Could not create CSV file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(fields); err != nil {
		log.Fatal("Error writing header to CSV:", err)
	}

	for _, record := range records {
		row := []string{}
		for _, field := range fields {
			value := ""
			if v, ok := record[field]; ok {
				switch v := v.(type) {
				case []interface{}:
					if len(v) >= 2 {
						value = fmt.Sprintf("%v", v[1])
					}
				default:
					value = fmt.Sprintf("%v", v)
				}
			}
			row = append(row, value)
		}
		if err := writer.Write(row); err != nil {
			log.Fatal("Error writing record to CSV:", err)
		}
	}

	fmt.Printf("%s have been exported to %s\n", fields, fileName)
}

func main() {
	getOdoo()
}
