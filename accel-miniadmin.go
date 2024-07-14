/*
ACCEL-PPP mini admin web interface
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"encoding/json"
	"bufio"
	"time"
)

var (
	Hostname string
	Admintoken string
	Chapfile string
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Redirect to /static/admin.html
	http.Redirect(w, r, "/static/admin.html", http.StatusFound)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	// set no-cache
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	// verify path
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// serve static files
	http.ServeFile(w, r, r.URL.Path[1:])
}

/*
{
	"action": "del/add/show",
	"username": "username",
	"password": "password",

	also token validation:
	"token": "SECRETTOKEN"
}
*/

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

/*
Add user to chap-secrets
username pptpd password 192.168.100.2-254
*/

func findFreeIP() (string, error) {
	// open /etc/ppp/chap-secrets
	var busyLastOctets []int
	file, err := os.Open(Chapfile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "192.168.100.") {
			// get last octet
			parts := strings.Fields(line)
			if len(parts) != 4 {
				continue
			}
			var lastOctet int
			n, err := fmt.Sscanf(parts[3], "192.168.100.%d", &lastOctet)
			if err != nil {
				return "", err
			}
			if n != 1 {
				continue
			}
			busyLastOctets = append(busyLastOctets, lastOctet)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	// find free IP
	for i := 2; i <= 254; i++ {
		if !contains(busyLastOctets, i) {
			return fmt.Sprintf("192.168.100.%d", i), nil
		}
	}
	return "", fmt.Errorf("No free IP addresses")
}

func isUserExists(username string) (bool, error) {
	file, err := os.Open(Chapfile)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// split by space/tab 4 parts
		parts := strings.Fields(line)
		if len(parts) != 4 {
			continue
		}
		if parts[0] == username {
			return true, nil
		}

	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	return false, nil
}

func addUser(username string, password string) (error, string) {
	// check if user exists
	exists, err := isUserExists(username)
	if err != nil {
		return err, ""
	}
	if exists {
		return fmt.Errorf("User already exists"), ""
	}

	// find free IP
	ip, err := findFreeIP()
	if err != nil {
		return err, ""
	}

	// add user to chap-secrets
	file, err := os.OpenFile(Chapfile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err, ""
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%s pptpd %s %s\n", username, password, ip)
	if err != nil {
		return err, ""
	}

	return nil, ip
}

func delUser(username string) error {
	// check if user exists
	exists, err := isUserExists(username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("User not found")
	}

	// remove user from chap-secrets
	file, err := os.OpenFile(Chapfile, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	newcontent := ""
	for scanner.Scan() {
		line := scanner.Text()
		// split by space/tab 4 parts
		parts := strings.Fields(line)
		if len(parts) != 4 {
			newcontent += line + "\n"
			continue
		}
		if parts[0] == username {
			// remove line
		} else {
			// write line
			newcontent += line + "\n"
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// write new content
	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = file.WriteString(newcontent)
	if err != nil {
		return err
	}

	return nil
}

type User struct {
	Username string `json:"username"`
	IP string `json:"ip"`
}

type UserList struct {
	Users []User `json:"users"`
	Status string `json:"status"`
}

func showUsers(w http.ResponseWriter, r *http.Request) {
	// Show username / ip only
	file, err := os.Open(Chapfile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	users := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		// split by space/tab 4 parts
		parts := strings.Fields(line)
		if len(parts) != 4 {
			continue
		}
		users[parts[0]] = parts[3]
	}
	if err := scanner.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*
	{ "status": "success",
	  "users": {
	  			"username1": "ip1",
				"username2": "ip2"
	  }
	}
	*/


	var userList UserList
	userList.Status = "success"
	for username, ip := range users {
		userList.Users = append(userList.Users, User{Username: username, IP: ip})
	}

	js, err := json.Marshal(userList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		// empty list
		js = []byte(`{"message": "success", "users": [], "status": "No users added yet"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	// OK
	w.WriteHeader(http.StatusOK)
}

func validateToken(w http.ResponseWriter, r *http.Request, token string) {
	if token != Admintoken {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}
	// json status = 'success', token = token
	js := fmt.Sprintf(`{"message": "success", "token": "%s", "status": "success"}`, token)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(js))
	// OK
	w.WriteHeader(http.StatusOK)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	jsonDecoder := json.NewDecoder(r.Body)
	var data map[string]string
	err := jsonDecoder.Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	// if parameter token exists, check it
	if data["token"] != "" {
		validateToken(w, r, data["token"])
		return
	}
	// check if Authorization header exists
	if r.Header.Get("Authorization") == "" {
		http.Error(w, "Authorization required", http.StatusForbidden)
		return
	}
	// check if Authorization header is valid
	if r.Header.Get("Authorization") != "Bearer " + Admintoken {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	if data["action"] == "show" {
		showUsers(w, r)
		return
	}

	if data["action"] == "" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}
	if data["username"] == "" {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}
	if data["action"] == "add" {
		err, ip := addUser(data["username"], data["password"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			js := fmt.Sprintf(`{"ip": "%s", "message": "success", "status": "success"}`, ip)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(js))
			// OK
			w.WriteHeader(http.StatusOK)
			return
		}
	} else if data["action"] == "delete" {
		err := delUser(data["username"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			// OK
			js := fmt.Sprintf(`{"message": "User deleted", "status": "success"}`)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(js))
			w.WriteHeader(http.StatusOK)
			return
		}
	} else {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}
}

/*
	Search certificates in /etc/letsencrypt/live and install them to /etc/ssl/
    ln -s /etc/letsencrypt/live/${SSTP_HOSTNAME}/fullchain.pem /etc/accel-ppp/ca.crt
    ln -s /etc/letsencrypt/live/${SSTP_HOSTNAME}/privkey.pem /etc/accel-ppp/server.key
    ln -s /etc/letsencrypt/live/${SSTP_HOSTNAME}/cert.pem /etc/accel-ppp/server.crt
*/
func installCerts() {
	// check if letsencrypt fullchain.pem exists, if not keep looping and sleeping
	for {
		log.Printf("Checking for /etc/letsencrypt/live/%s/fullchain.pem", Hostname)
		if _, err := os.Stat(fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", Hostname)); err == nil {
			break
		}
		log.Printf("File not found, sleeping 5 seconds, seems like certbot is still running")
		time.Sleep(5 * time.Second)
	}

	// is file already exists?
	if _, err := os.Stat("/etc/ssl/ca.crt"); err != nil {
		err := os.Symlink(fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", Hostname), "/etc/ssl/ca.crt")
		if err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat("/etc/ssl/server.key"); err != nil {
		err := os.Symlink(fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", Hostname), "/etc/ssl/server.key")
		if err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat("/etc/ssl/server.crt"); err != nil {
		err := os.Symlink(fmt.Sprintf("/etc/letsencrypt/live/%s/cert.pem", Hostname), "/etc/ssl/server.crt")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	var err error
	var port string
	var privkey string
	var cert string
	var cacert string
	var httponly string
	

	flag.StringVar(&port, "port", "8080", "port to listen on")
	flag.StringVar(&privkey, "privkey", "/etc/ssl/server.key", "path to private key")
	flag.StringVar(&cert, "cert", "/etc/ssl/server.crt", "path to certificate")
	flag.StringVar(&cacert, "cacert", "/etc/ssl/ca.crt", "path to CA certificate")
	flag.StringVar(&Chapfile, "chapfile", "/etc/ppp/chap-secrets", "path to chap-secrets file")
	// force http
	flag.StringVar(&httponly, "http", "false", "force http")

	flag.Parse()

	// Get hostname from env SSTP_HOSTNAME
	Hostname = os.Getenv("SSTP_HOSTNAME")
	if Hostname == "" {
		log.Fatal("SSTP_HOSTNAME not set")
	}

	// Get admin password from env SSTP_ADMINTOKEN
	Admintoken = os.Getenv("SSTP_ADMINTOKEN")
	if Admintoken == "" {
		log.Fatal("SSTP_ADMINTOKEN not set")
	}

	installCerts()

	http.HandleFunc("/", handler)
	http.HandleFunc("/static/", staticHandler)
	//http.HandleFunc("/restart", restartHandler)
	//http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/api", apiHandler)

	// check if SSL is forced
	if httponly == "true" {
		log.Printf("Listening on port %s", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}


	// check SSL certificates
	if _, err = os.Stat(privkey); os.IsNotExist(err) {
		log.Fatalf("Private key file %s not found", privkey)
	}
	if _, err = os.Stat(cert); os.IsNotExist(err) {
		log.Fatalf("Certificate file %s not found", cert)
	}
	if _, err = os.Stat(cacert); os.IsNotExist(err) {
		log.Fatalf("CA certificate file %s not found", cacert)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServeTLS(":"+port, cert, privkey, nil))
}
