package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type PassthroughWriter struct {
	http.ResponseWriter
	Status int
}

func (w *PassthroughWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *PassthroughWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *PassthroughWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

type TemplateParser func(mux *http.ServeMux, path, template string) bool

func StatusCodeTemplateParser(mux *http.ServeMux, path, template string) bool {
	status, err := strconv.Atoi(template)
	if err != nil {
		return false
	}
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(http.StatusText(status)))
	})
	return true
}

func DirTemplateParser(mux *http.ServeMux, path, template string) bool {
	info, err := os.Stat(template)
	if (err != nil && os.IsNotExist(err)) || !info.IsDir() {
		return false
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	mux.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir(template))))
	return true
}

func FileTemplateParser(mux *http.ServeMux, path, template string) bool {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, template)
	})
	return true
}

func logHandler(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writer := &PassthroughWriter{ResponseWriter: w}
		handler.ServeHTTP(writer, r)
		log.Printf("- %s - %s %s - %d\n", r.RemoteAddr, r.Method, r.URL, writer.Status)
	}
}

func usageFunc() {
	fmt.Println("USAGE:")
	fmt.Println("  mhs [options] [/request-path response-template]...")
	fmt.Println()
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("RESPONSE TEMPLATES")
	fmt.Println("A response template can either be a status code, a path to an existing directory, or a file. It is assumed to be a file path if it is not a valid status code and does not exist as a directory.")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println("Serve current directory on port 8080:")
	fmt.Println("  mhs -p 8081")
	fmt.Println("Serve current directory on port 8081:")
	fmt.Println("  mhs -p 8081")
	fmt.Println("Serve 200s from /ok and 500s from /error:")
	fmt.Println("  mhs /ok 200 /error 500")
	fmt.Println("Serve 200s from /status and the \"/tmp\" directory from /files:")
	fmt.Println("  mhs /status 200 /files /tmp")
}

func main() {
	port := flag.Int("p", 8080, "port to serve on")
	flag.Usage = usageFunc
	flag.Parse()

	log.Default().SetFlags(log.Ldate | log.Lmicroseconds)
	args := flag.Args()
	if len(args) == 0 {
		fmt.Printf("Serving current directory via HTTP on port %d\n", *port)
		http.ListenAndServe(fmt.Sprintf(":%d", *port), logHandler(http.FileServer(http.Dir("."))))
		return
	}

	if len(args) % 2 != 0 {
		fmt.Println("Please specify pairs of paths and response templates")
		return
	}

	templateParsers := []TemplateParser{
		StatusCodeTemplateParser,
		DirTemplateParser,
		FileTemplateParser,
	}

	mux := http.NewServeMux()
	for i := 0; i < len(args); i += 2 {
		path := args[i]
		responseTemplate := args[i + 1]
		for _, parser := range templateParsers {
			if parser(mux, path, responseTemplate) {
				break
			}
		}
	}

	log.Printf("Serving HTTP on port %d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), logHandler(mux)))
}
