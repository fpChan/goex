package monitor

type exchange interface {
	getPrice(symbol string) float64
}

type Monitor interface {
	Start() error
}
