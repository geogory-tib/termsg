package serverstate

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"server/consts"
	"server/types"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type connection_manager struct {
	mu        sync.RWMutex
	act_users map[string]net.Conn
}
type user_manager struct {
	mu    sync.RWMutex
	users map[string][32]byte
}
type Server_state struct {
	active_user   connection_manager
	users         user_manager
	Ip            string `json:"ip"`
	Port          string `json:"port"`
	Login_message string `json:"loginmessage"`
	listener      net.Listener
}

func Init() (ret *Server_state) {
	ret = new(Server_state)
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	config_json, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(config_json, &ret)
	ret.listener, err = net.Listen("tcp", ret.Ip+":"+ret.Port)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func (self *Server_state) Server_Main() {
	for {
	start:
		conn, err := self.listener.Accept()
		if err != nil {
			conn.Close()
			log.Println(err)
			goto start
		}
		_, err = ws.Upgrade(conn)
		if err != nil {
			conn.Close()
			log.Println(err)
			goto start
		}
		go self.handle_func(conn)
	}
}

func (self *Server_state) handle_func(conn net.Conn) {
	defer conn.Close()
	for pass := false; pass; pass = self.login_user(conn) {
	}
	message_strct := types.Message{}
	for {
		client_buf, err := wsutil.ReadClientText(conn)
		if err != nil {
			log.Println(err)
			return
		}
		json.Unmarshal(client_buf, &message_strct)
		switch message_strct.Type {
		case consts.COMMAND:
			self.handle_command(conn, message_strct)

		case consts.MESSAGE:
			self.handle_message(conn, message_strct)
		}

	}
}

func (self *Server_state) login_user(conn net.Conn) bool {
	message := types.Message{
		Time:     time.Now().Format(time.ANSIC),
		From:     "server",
		Contents: self.Login_message,
		Type:     consts.LOGIN,
	}
	loginbuff, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return false
	}
	err = wsutil.WriteServerText(conn, loginbuff)
	if err != nil {
		log.Println(err)
		return false
	}
	loginbuff, err = wsutil.ReadClientText(conn)
	if err != nil {
		log.Println(err)
		return false
	}
	login_info := types.User{}
	json.Unmarshal(loginbuff, &login_info)
	self.users.mu.RLock()
	password_sum := sha256.Sum256(login_info.Password[:]) // i know this isn't the most secure but I will switch to bcrypt or something later
	defer self.users.mu.RUnlock()
	user_pass, is_a_user := self.users.users[login_info.Name]
	if !is_a_user {
		self.users.mu.RUnlock()
		self.users.mu.Lock()
		defer self.users.mu.Unlock()
		self.users.users[login_info.Name] = password_sum
		return true
	}
	if password_sum == user_pass {
		self.active_user.mu.Lock()
		defer self.active_user.mu.Unlock()

		self.active_user.act_users[login_info.Name] = conn
	}
	return true
}

func (self *Server_state) Desconstruct() {
	self.active_user.mu.Lock()
	for _, conn := range self.active_user.act_users {
		conn.Close()
	}
	self.listener.Close()
}

func (self *Server_state) handle_command(conn net.Conn, command_message types.Message) {
	split_str := strings.Split(command_message.Contents, " ")
	switch split_str[0] {
	case consts.SHOW_COMMAND:
		switch split_str[1] {
		case consts.SHOW_ACTIVE:
			builer := strings.Builder{}
			builer.Grow(100)
			self.active_user.mu.RLock()
			defer self.active_user.mu.RUnlock()
			for user := range self.active_user.act_users {
				builer.WriteString(user + "\n")
			}
			reply_message := types.Message{
				Type:     consts.COMMAND_REPLY,
				From:     "server",
				Contents: builer.String(),
				Time:     time.Now().Format(time.ANSIC),
			}
			json_buffer, err := json.Marshal(reply_message)
			if err != nil {
				log.Println(err)
				return
			}
			err = wsutil.WriteServerText(conn, json_buffer)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (self *Server_state) handle_message(conn net.Conn, msg_struct types.Message) {
	username_to_send := msg_struct.To
	self.active_user.mu.Lock()
	defer self.active_user.mu.Unlock()
	user_to_send_to, isuseractive := self.active_user.act_users[username_to_send]
	if isuseractive {
		msg_json, err := json.Marshal(msg_struct)
		if err != nil {
			log.Println(err)
			self.send_message_error_to_client(conn, username_to_send)
		}
		err = wsutil.WriteServerText(user_to_send_to, msg_json)
		if err != nil {
			self.send_message_error_to_client(conn, username_to_send)
		}
	} else {
		user_not_exist_msg := types.Message{
			Type: consts.MESSAGE_FAIL,
			From: "server",
			Time: time.Now().Format(time.ANSIC),
		}
		user_not_exist_json, err := json.Marshal(user_not_exist_msg)
		if err != nil {
			log.Println(err)
		}
		err = wsutil.WriteServerText(conn, user_not_exist_json)
		if err != nil {
			log.Println(err)
		}
	}
}

func (self *Server_state) send_message_error_to_client(conn net.Conn, username string) {
	failure_msg := types.Message{
		Type:     consts.MESSAGE_FAIL,
		From:     "server",
		Contents: "Failure sending message to " + username,
		Time:     time.Now().Format(time.ANSIC),
	}
	err_msg, err := json.Marshal(failure_msg)
	if err != nil {
		log.Println(err)
		return
	}
	err = wsutil.WriteServerText(conn, err_msg)
	if err != nil {
		log.Println(err)
		return
	}
}
