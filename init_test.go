package proxman

import "net/url"

var (
	p1, p2, p3, p4, p5 *url.URL
)

func init() {
	p1, _ = url.Parse("http://127.0.0.1:8080")
	p2, _ = url.Parse("http://127.0.0.1:8081")
	p3, _ = url.Parse("http://127.0.0.2:8080")
	p4, _ = url.Parse("http://127.0.0.2:8081")
	p5, _ = url.Parse("http://127.0.0.3:8080")
}
