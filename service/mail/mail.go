package mail

import (
	"log"
	"errors"
	//"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
  "github.com/curious-eyes/jmail"
	database "titable.go/db"
)


// 授業のclass構造体とメールアドレスが渡されて，その授業のURLを返す．
func GetURL(class database.Class, user database.User) ([]string, error){
	//メールからZOOMのURLをとってきて返す
	c, err := login(user.EmailID, user.UserAuth)
	if err != nil {
		return nil, err
	}
	defer c.Logout() //必ずlogout

	//iヶ月分の受信メールを読む
	urls := readInBox(c, class, user)

	return urls, nil
}

func login(mailID string, pass string) (*client.Client, error){
	c, err := client.DialTLS("mailv3.m.titech.ac.jp:993", nil) //titech mail IMAP receive server
	if err != nil {
		return nil, errors.New("can't connect to mail server")
	}
	if err := c.Login(mailID, pass); err != nil {
		return nil, errors.New("can't login mail server")
	}
	return c, nil
}

//INBOXを見てメールを読む
func readInBox(c *client.Client, class database.Class, user database.User) []string {
	mbox, err := c.Select("INBOX", true)
	if err != nil {
		log.Fatal(err)
	}

	if mbox.Messages == 0 {
		log.Fatal("No message in mailbox")
	}

	criteria := imap.NewSearchCriteria()
	criteria.Since = user.LastTime
	criteria.Text = append(criteria.Text, class.Class)
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}
	
	var urls []string

	if len(ids) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)

		section := &imap.BodySectionName{}
		items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}

		messages := make(chan *imap.Message)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, items, messages)
		}()

		for msg := range messages{
			if msg == nil {
				log.Fatal("Server didn't returned message")
			}

			r := msg.GetBody(section)
			if r == nil {
				log.Fatal("Server didn't returned message body")
			}

			m, err := jmail.ReadMessage(r)
			if err != nil {
				log.Fatal(err)
			}

			body, err := m.DecBody()
			if err != nil {
				log.Fatal(err)
			}
			
			url, err := FetchURL(string(body))
			if err == nil {
				urls = append(urls, url)
			}
		} 
	}

	return urls
}