package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter , r *http.Request){
		// create a deferred function which will run constantly while unwinding error stack
		defer func ()  {
			if err := recover(); err != nil {
				w.Header().Set("Connection" , "close") // close the connection once error occurs
				app.serverErrorResponse(w , r , fmt.Errorf("%s" , err))
			}
		}()

		next.ServeHTTP(w , r)
	})
}

/*--------------------------
	1️⃣ Global rate limiting

func (app *application) rateLimit(next http.Handler) http.Handler{

limiter := rate.NewLimiter(2 , 4) // bucket size of 4 with refilling rate of 2tokens/s

// return a closure functions , it closes over the limiter variable defined above
return http.HandlerFunc(func(w http.ResponseWriter , r *http.Request){
	if !limiter.Allow() {
		app.rateLimitExceededResponse(w , r)
		return
	}

	next.ServeHTTP(w , r)
})

}

*/


// IP-based rate-limiting
func (app *application) rateLimit(next http.Handler) http.Handler{

	// define a client struct to hold the rate limiter and last seen time
	type client struct {
		limiter rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex // centralized mutex for rate-limiting purpose
		clients = make(map[string]*client)
	)

	// background goroutine for cleaning up old entries (per minute check and every 3/mins of inconsistency , cleanup)
	go func(){
		// infinite loop
		for {
			time.Sleep(time.Minute)

			// lock the mutex while the clean
			mu.Lock()

			for ip , client := range clients{
				if time.Since(client.lastSeen) > 3 * time.Minute {
					delete(clients , ip)
				}
			}

			mu.Unlock()
		}
	}() //

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retrieve the IP address from the request
		ip := realip.FromRequest(r)

		// lock the mutex to prevent race conditions by concurrent code
		mu.Lock()

		// check if the client exists in the map and if not add their IP with a rate limiter instance
		if _ , found := clients[ip]; !found {
			clients[ip] = &client{limiter: *rate.NewLimiter(2, 4)}
		}

		// once the limiter been assigned , attach current time
		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow(){
			mu.Unlock()
			app.rateLimitExceededResponse(w , r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w , r)

	})

}