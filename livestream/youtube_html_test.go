package livestream_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
)

func Test1(tt *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/live01.html")
		if err != nil {
			tt.Fatal(err)
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		fmt.Fprintln(w, string(b))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_, err := livestream.GetUpcommingLiveID(u)
	if err != nil {
		tt.Fatal(err)
	}
}

func Test2(tt *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/live02_none.html")
		if err != nil {
			tt.Fatal(err)
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		fmt.Fprintln(w, string(b))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_, err := livestream.GetUpcommingLiveID(u)
	if err != nil {
		tt.Fatal(err)
	}
}

func Test3(tt *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/live03_other.html")
		if err != nil {
			tt.Fatal(err)
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		fmt.Fprintln(w, string(b))
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	_, err := livestream.GetUpcommingLiveID(u)
	if err == nil {
		tt.Fatal("unexpected result")
	}
}

func Test4(tt *testing.T) {
	// now := time.Now().Add(time.Hour * 3)
	now := time.Now()
	// start := time.Unix(1575889200, 0) // 2019/12/09 20:00
	start := time.Unix(1575900000, 0) // 2019/12/09  22:00

	ok := start.After(now)
	fmt.Println(fmt.Sprintf("sub:%v", start.Sub(now)))
	if ok && start.Sub(now) < (time.Hour*6) {
		return
	}
	tt.Error("error")
}

func Test5(tt *testing.T) {
	checkAndStartSubscribedChannel("")
}

var cnt = 0

// CheckAndStartSubscribedChannel rrr
func checkAndStartSubscribedChannel(nextPageToken string) {
	fmt.Println("checkAndStartSubscribedChannel")
	cnt++
	NextPageToken := ""
	if cnt < 5 {
		NextPageToken = fmt.Sprintf("%v", cnt)
	}

	if NextPageToken != "" {
		checkAndStartSubscribedChannel(NextPageToken)
	}
}
