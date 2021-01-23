package main

// Based on https://github.com/schwarmco/go-cartesian-product
// but single-threaded with reproducible output

// Iter takes interface-slices and returns a channel, receiving cartesian products
func Iter(params ...[]interface{}) chan []interface{} {
	c := make(chan []interface{})
	go func() {
		iterate(c, []interface{}{}, params...)
		close(c)
	}()
	return c
}

func iterate(channel chan []interface{}, result []interface{}, params ...[]interface{}) {
	if len(params) == 0 {
		channel <- result
		return
	}
	p, params := params[0], params[1:]
	for i := 0; i < len(p); i++ {
		resultCopy := append([]interface{}{}, result...)
		iterate(channel, append(resultCopy, p[i]), params...)
	}
}
