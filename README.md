# Go Techan/Interactive Brokers Adapter

translation library for interacting with interactive brokers using the tech go library. this go package provides
lightweight convenience functions for pulling data from the interactive brokers web api and converting the responses
to structs found in the techan package. this allows trading algo and data implementations to easily operate on live
data from an interactive brokers trading account.

links to both the techan and interactive brokers web api packages can be found below:

| package                     | link                                                                                           |
|-----------------------------|------------------------------------------------------------------------------------------------|
| techan                      | [https://github.com/schmidthole/techan](https://github.com/schmidthole/techan)                 |
| interactive brokers web api | [https://github.com/schmidthole/ibkr-webapi-go](https://github.com/schmidthole/ibkr-webapi-go) |

this package serves my purposes as a lightweight library which can live externally from actual algo/data processing implementations.