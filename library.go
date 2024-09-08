package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Book struct {
	ID          int
	Title       string
	Author      string
	PublishDate string
	InDate      string
	IsBorrowed  bool
	Borrower    string
}

var books []Book
var token string
var currentUsername string

var mySigningKey = []byte("HakimYu") // 用于签名的密钥
// 生成 Token
func generateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // 1小时后过期
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

// 验证 Token
func checkToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["username"].(string), nil
	} else {
		return "", err
	}
}

// 用户结构体
type User struct {
	Username string
	Password string
}

// 读取用户数据
func readUsersFromJSON() []User {
	data, err := os.ReadFile("users.json")
	if err != nil {
		return nil
	}
	var users []User
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil
	}
	return users
}

// 注册用户
func register() {
	scanner := bufio.NewScanner(os.Stdin)

	// 输入用户名
	fmt.Print("请输入用户名: ")
	scanner.Scan()
	username := scanner.Text()
	username = strings.TrimSpace(username) // 去除空白字符

	// 输入密码
	fmt.Print("请输入密码: ")
	scanner.Scan()
	password := scanner.Text()
	password = strings.TrimSpace(password) // 去除空白字符

	// 验证用户名是否已存在
	users := readUsersFromJSON()
	for _, user := range users {
		if user.Username == username[:len(username)-1] {
			fmt.Println("用户名已存在,请重新输入。")
			return
		}
	}

	// 注册用户
	newUser := User{
		Username: username[:len(username)-1],
		Password: password[:len(password)-1],
	}
	users = append(users, newUser)

	// 将 users 写入 JSON 文件
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		fmt.Println("写入 JSON 文件失败:", err)
	}
	err = os.WriteFile("users.json", data, 0644)
	if err != nil {
		fmt.Println("写入 JSON 文件失败:", err)
	}
	fmt.Println("注册成功！")
	generateToken(username)
}

func addBook() {
	_,err := checkToken(token)

	if err!= nil {
		fmt.Println("请登录！", err)
		login()
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	books = readBooksFromJSON()
	fmt.Println("请输入图书标题:")
	scanner.Scan()
	title := scanner.Text()

	fmt.Println("请输入作者:")
	scanner.Scan()
	author := scanner.Text()

	fmt.Println("请输入出版日期:")
	scanner.Scan()
	publishDate := scanner.Text()

	fmt.Println("请输入入库日期:")
	scanner.Scan()
	inDate := scanner.Text()

	newBook := Book{
		Title:       title,
		Author:      author,
		PublishDate: publishDate,
		InDate:      inDate,
		IsBorrowed:  false,
	}
	if len(books) == 0 {
		newBook.ID = 1
	} else {
		newBook.ID = books[len(books)-1].ID + 1
	}
	books = append(books, newBook)
	fmt.Println("图书添加成功！")

	// 将 books 写入 JSON 文件
	err = writeBooksToJSON()
	if err != nil {
		fmt.Println("写入 JSON 文件失败:", err)
	}
}
func deleteBook() {
	_,err := checkToken(token)

	if err!= nil {
		fmt.Println("请登录！", err)
		login()
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	// 实现删除图书的逻辑
	fmt.Println("请输入要删除的图书 ID:")
	scanner.Scan()
	idStr := scanner.Text()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("输入错误,请重新输入。")
		return
	}
	for i, book := range books {
		if book.ID == id {
			books = append(books[:i], books[i+1:]...)
			fmt.Println("图书删除成功！")

			// 将 books 写入 JSON 文件
			err := writeBooksToJSON()
			if err != nil {
				fmt.Println("写入 JSON 文件失败:", err)
			}
			return
		}
	}
	fmt.Println("未找到该图书 ID。")
}
func queryBook() {
	// 实现查询图书的逻辑
	fmt.Println("当前图书列表:")
	books = readBooksFromJSON()
	for _, book := range books {
		fmt.Printf("ID: %d, 标题: %s, 作者: %s, 出版日期: %s, 入库日期: %s, 是否借出: %t, 借阅人: %s\n",
			book.ID, book.Title, book.Author, book.PublishDate, book.InDate, book.IsBorrowed, book.Borrower)
	}
}
func borrowBook() {
	_,err := checkToken(token)

	if err!= nil {
		fmt.Println("请登录！", err)
		login()
		return
	}
	fmt.Println("请输入要借阅的图书 ID:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	idStr := scanner.Text()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("输入错误,请重新输入。")
		return
	}

	for i, book := range books {
		if book.ID == id {
			if book.IsBorrowed {
				fmt.Println("该图书已被借阅。")
			} else {
				borrower := currentUsername
				books[i].Borrower = borrower
				books[i].IsBorrowed = true
				fmt.Println("图书借阅成功！")

				// 将 books 写入 JSON 文件
				err := writeBooksToJSON()
				if err != nil {
					fmt.Println("写入 JSON 文件失败:", err)
				}
			}
			return
		}
	}
	fmt.Println("未找到该图书 ID。")
}
func returnBook() {
	// 实现归还图书的逻辑
	fmt.Println("请输入要归还的图书 ID:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	idStr := scanner.Text()
	id, _ := strconv.Atoi(idStr)

	for i, book := range books {
		if book.ID == id {
			if !book.IsBorrowed {
				fmt.Println("该图书未被借阅。")
			} else {
				books[i].IsBorrowed = false
				books[i].Borrower = ""
				fmt.Println("图书归还成功！")

				// 将 books 写入 JSON 文件
				err := writeBooksToJSON()
				if err != nil {
					fmt.Println("写入 JSON 文件失败:", err)
				}
			}
			return
		}
	}
	fmt.Println("未找到该图书 ID。")
}
func writeBooksToJSON() error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("books.json", data, 0644)
}
func readBooksFromJSON() []Book {
	data, err := os.ReadFile("books.json")
	if err != nil {
		return nil
	}
	var books []Book
	err = json.Unmarshal(data, &books)
	if err != nil {
		return nil
	}
	return books
}
func login() {
	scanner := bufio.NewScanner(os.Stdin)

	// 输入用户名
	fmt.Print("请输入用户名: ")
	scanner.Scan()
	username := scanner.Text()
	// 输入密码
	fmt.Print("请输入密码: ")
	scanner.Scan()
	password := scanner.Text()

	// 验证用户
	if check(username, password) {
		fmt.Println("登录成功！")
		tokenStr, err := generateToken(username) 
		if err != nil {
			fmt.Println("生成 Token 失败:", err)
		} else {
			token = tokenStr
			currentUsername = username
		}
	} else {
		fmt.Println("用户名或密码错误。")
	}
}

func check(username, password string) bool {
	users := readUsersFromJSON()
	for _, user := range users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}
func main() {
	for {
		fmt.Println("请输入操作类别(1:添加图书, 2:删除图书,3:查询图书,4: 借阅图书,5: 归还图书,6: 登录, 7: 注册,   输入其它:退出）:")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("输入错误,请重新输入。")
			break
		}

		switch choice {
		case 1:
			addBook()
		case 2:
			deleteBook()
		case 3:
			queryBook()
		case 4:
			borrowBook()
		case 5:
			returnBook()
		case 6:
			login()
		case 7:
			register()
		default:
			fmt.Println("退出程序。")
			return
		}
	}
}
