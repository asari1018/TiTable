# TiTable

## UI周り仕様
| 画面 | htmlファイル |
| :-----: | :-----: |
|初期画面 | index.html |
|ログイン画面 | login.html|
|アカウント登録画面 | signup.html|
|(カレンダーのある)メイン画面 | main.html|
|タスク詳細画面 | task.html|
|授業詳細画面 | class.html|
|アカウント編集画面 | account.html|



### 初期画面（index.html)
* 呼び出し元:　
  home.go/Home 
* 受け取る情報:
  ```gin.H{"Title": "HOME"}```
* サーバでの行き先(PATH): 
    * account.go/Login (```/login```)
    * account.go/Signup (```/signup```)
* サーバに送る情報:なし


### ログイン画面（login.html)
* 呼び出し元:　
  home.go/LoginPage
  account.go/Signout(サインアウト時)
  account.go/Login（ログイン失敗時）
* 受け取る情報:
  ```gin.H{"Title": "LOGIN"}```
  ```gin.H{"Title": "LOGIN", "Info": "ユーザネームまたはパスワードが誤っています»}```（ログイン失敗時）
* サーバでの行き先(PATH): account.go/Login (```/login```)
* サーバに送る情報: ユーザー名　パスワード
  送り方: formタグのPOSTリクエスト
ユーザ名(key=user)　パスワード(key=pw)





### アカウント登録画面(signup.html)
* 呼び出し元: 
    * home.go/SignupPage　
    * account.go/Signup （登録失敗時）
* 受け取る情報: 　
    * ```gin.H{"Title": "SIGNUP"}```
    * ```gin.H{"Title": "SIGNUP", "InfoMail": "すでに使用されているメールアドレスです"}```（登録失敗時）
    * ```gin.H{"Title": "SIGNUP", "InfoUser": "すでに使用されているアカウント名です"}```（登録失敗時）
* サーバでの行き先(PATH):
    * account.go/Signup (```/signup```)
* サーバに送る情報:　
  * メールアドレス　(自分で作った)パスワード　ユーザ名　ユーザ認証情報（送り方: formタグのPOSTリクエスト）
     メールID(key=email_id)　パスワード(key=pw)　ユーザ名(key=user) ユーザ認証情報(user_auth)
  * 　iカレンダー


### (カレンダーのある)メイン画面(main.html)
* 呼び出し元 :　titable.go/Main
* 受け取る情報 :
    * ```gin.H{"Title": "TITABLE", "Tasks": tasks, "Classes": classes}```
    * ```tasks []Task``` (db/schema.go参照)　←Task構造体の配列
    * ```classes []Class``` (db/schema.go参照)　←Class構造体の配列
* サーバでの行き先（1）: タスク詳細画面
    *  class.go/Task  
	* サーバに送る情報: 選択されたタスク名
　送り方: Pathパラメータ　urlの「/main/task/:task」に合うように
* サーバでの行き先（2）: 授業詳細画面
    *  class.go/Class  
	* サーバに送る情報: 選択された授業名
　送り方: Pathパラメータ　urlの「/main/class/:class」に合うように

* サーバでの行き先(PATH)（3）: アカウント編集画面
    * account.go/AccountEdit (```/account```)  
	* サーバに送る情報: なし

* サーバでの行き先(PATH)(4) : ログアウトボタンから
    * account.go/Signout (```/signout```)
	* サーバに送る情報: なし

### 授業詳細画面(class.html)
* 呼び出し元 :　class.go/Class
* 受け取る情報 :
    * ```　gin.H{"Title": "Class", "Tasks": tasks, "Class": class}```
    * ```tasks []Task``` (db/schema.go参照) ←こっちはTask構造体の配列
    * ```class Class``` (db/schema.go参照)　←こっちはただのClass構造体
* サーバでの行き先(PATH): class.go/TaskInsert(```/main/class/taskinsert```)
* サーバに送る情報: 
    * タスク名(key=title) 

### タスク詳細画面(task.html)
* 呼び出し元 :  task.go/Task
* 受け取る情報 :
    * ```gin.H{"Title": "TASK", "Task": task}```
    * ```task Task``` (db/schema.go参照)　←Task構造体
* サーバでの行き先(1)(PATH):　task.go/TaskEdit (```/main/taskedit```)
* サーバに送る情報: 　
    * タスク名 (key=title)
    * タスクレベル (key=level)
    * 締切時間 (key=deadline)
      (type=datetime-local?)
      https://developer.mozilla.org/ja/docs/Web/HTML/Element/input/datetime-local
* サーバでの行き先(2)(PATH): task.go/TaskDone (完了ボタンからくる) (```/main/taskdone/:task```)
* サーバに送る情報:
    * タスク名
　送り方: Pathパラメータ　urlの「/main/taskdone/:task」に合うように


### アカウント編集画面(account.html)
* 呼び出し元 :　account.go/AccountEdit
* 受け取る情報 :
    * ```　gin.H{"Title": "ACCOUNT", "User": user}```
    * ```user User``` (db/schema.go参照)　←User構造体
* サーバでの行き先(PATH): account.go/AccountEdit
* サーバに送る情報: 
    * メールID(key=email_id)
    * パスワード(key=pw)
    * ユーザネーム(key=user)


    ## How to run the application
First, you need to start Docker containers.
```sh
$ docker-compose up -d
```
This command will take time to download and build the containers.

Now you can start the application with the following command.
```sh
$ docker-compose exec app go run main.go
```
You can also execute `go run main.go` directly if you have Go development tools on your machine, but you need to setup the configuration to connect the application with MySQL server.

When you finish exercise, please don't forget to stop the containers.
```sh
$ docker-compose down
```

## Advanced: How to initialize database
When you modify the database schema, you will need to discard the current DB volumes for creating a new one.
It will be easier to rebuild everything than to rebuild only DB container.
Following command helps you to do it.
```sh
$ docker-compose down --rmi all --volumes --remove-orphans
$ docker-compose up -d
```

