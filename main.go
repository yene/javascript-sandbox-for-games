package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/robertkrimen/otto"
)

var addr = flag.String("addr", ":8081", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var worldState string // JSON of the world state

type PayloadRequest struct {
	Type string          `json:"type"` // "world" or "code"
	Data json.RawMessage `json:"data"` // contains the bytes of the JSON, can be converte to string
	Code string          `json:"code"`
}

type PayloadResponse struct {
	Type   string          `json:"type"`   // Always "response"
	Status string          `json:"status"` // "ok" or "error"
	Error  string          `json:"error"`
	Data   json.RawMessage `json:"data"`
}

func sandbox(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// TODO: is this block a new go routine for each connection?
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// log.Printf("recv: %s", message)

		var request PayloadRequest
		err = json.Unmarshal(message, &request)
		if err != nil {
			log.Println("write:", err)
			break
		}
		if request.Type == "world" {
			log.Println("received world state")
			worldState = string(request.Data)
		}
		if request.Type == "code" {
			log.Println("received code, running it")
			vm := otto.New()
			vm.Set("world", worldState)
			_, err := vm.Run("world = JSON.parse(world)") // TODO: error handling, invalid JSON if this fails
			_, err = vm.Run(request.Code)
			if err != nil {
				log.Println("error in JS", err)
			}
			value, err := vm.Run("JSON.stringify(world)")
			if err != nil {
				log.Println("error in JS", err)
			}
			valueS, _ := value.ToString()
			// log.Println("value of world", value)

			resp := PayloadResponse{Type: "response", Status: "ok", Data: json.RawMessage(valueS)}
			respBytes, _ := json.Marshal(resp)
			err = c.WriteMessage(mt, respBytes)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/sandbox", sandbox)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
