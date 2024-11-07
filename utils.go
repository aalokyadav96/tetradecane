package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	rndm "math/rand"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var tmpl = template.Must(template.ParseGlob("index.html"))

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func generateID(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rndm.Intn(len(letters))]
	}
	return string(b)
}

// Helper function to remove a string from a slice
func removeString(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...) // Remove element
		}
	}
	return slice
}

// func uploadFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	if r.Method == http.MethodPost {
// 		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
// 		if err != nil {
// 			http.Error(w, "Unable to parse form", http.StatusBadRequest)
// 			return
// 		}

// 		file, fileHeader, err := r.FormFile("file")
// 		if err != nil {
// 			http.Error(w, "Unable to get file", http.StatusBadRequest)
// 			return
// 		}
// 		defer file.Close()

// 		// Ensure uploads directory exists
// 		if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
// 			http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
// 			return
// 		}

// 		out, err := os.Create(filepath.Join("./uploads", fileHeader.Filename))
// 		if err != nil {
// 			http.Error(w, "Unable to create file", http.StatusInternalServerError)
// 			return
// 		}
// 		defer out.Close()

// 		if _, err = io.Copy(out, file); err != nil {
// 			http.Error(w, "Unable to save file", http.StatusInternalServerError)
// 			return
// 		}

// 		fmt.Fprintf(w, "File uploaded successfully!")
// 	} else {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 	}
// }

//==================

// func sendImageAsBytes(w http.ResponseWriter, _ *http.Request, a httprouter.Params) {
// 	buf, err := os.ReadFile("./images/" + a.ByName("imageName"))
// 	if err != nil {
// 		log.Print(err)
// 	}
// 	w.Header().Set("Content-Type", "image/png")
// 	w.Write(buf)
// }

func CSRF(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, GenerateName(8))
}

func GenerateName(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rndm.Intn(len(letters))]
	}
	return string(b)
}

// func renderError(w http.ResponseWriter, message string, statusCode int) {
// 	w.WriteHeader(statusCode)
// 	w.Write([]byte(message))
// }

func EncrypIt(strToHash string) string {
	data := []byte(strToHash)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// Helper function to send responses
// func sendResponse(w http.ResponseWriter, status int, data interface{}, message string, err error) {
// 	response := Response{
// 		Message: message,
// 		Data:    data,
// 	}

// 	if err != nil {
// 		response.Error = err.Error()
// 	}

// 	w.WriteHeader(status)
// 	json.NewEncoder(w).Encode(response)
// }

func sendResponse(w http.ResponseWriter, status int, data interface{}, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"status":  status,
		"message": message,
		"data":    data,
	}

	if err != nil {
		response["error"] = err.Error()
	}

	// Encode response and check for encoding errors
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// func sendResponse(w http.ResponseWriter, status int, data interface{}, message string, err error) {
// 	w.WriteHeader(status)
// 	response := map[string]interface{}{
// 		"status":  status,
// 		"message": message,
// 		"data":    data,
// 	}
// 	if err != nil {
// 		response["error"] = err.Error()
// 	}
// 	json.NewEncoder(w).Encode(response)
// }
