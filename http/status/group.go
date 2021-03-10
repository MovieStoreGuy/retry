package status

const (
	_ int = iota
	Group1xx
	Group2xx
	Group3xx
	Group4xx
	Group5xx
)

var (
	text = map[int]string{
		Group1xx: "1xx",
		Group2xx: "2xx",
		Group3xx: "3xx",
		Group4xx: "4xx",
		Group5xx: "5xx",
	}
)

func Group(status int) int {
	return status / 100
}

func GroupText(group int) string {
	if s, ok := text[group]; ok {
		return s
	}
	return "?xx"
}
