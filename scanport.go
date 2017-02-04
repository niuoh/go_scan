package main

import (
	"fmt"
	"flag"
	"strings"
	"strconv"
	"net"
	"sync"
	"io/ioutil"
	"net/http"
	"regexp"
)

var (
	pl=fmt.Println
)

type Scan struct{
	ip_ string
	port_ int
	html_ bool
	model int
	ch chan int
	thread_num int
	port_list struct{
		sync.RWMutex
		m map[int]int
	}
	ip_list struct{
		sync.RWMutex
		m map[int]string
	}
}

func (s *Scan) get_title(ip string,port int) string{
	resp, err := http.Get("http://"+ip+":"+strconv.Itoa(port))
    if err != nil {
        return ""
    }
    defer resp.Body.Close()
    body,_ := ioutil.ReadAll(resp.Body)
    reg := regexp.MustCompile("<title>(.*)</title>")

    title := reg.Find([]byte(string(body)))
    if len(title)==0{
    	return ""
    }
    return string(title)
}

func (s *Scan) connect(host string,port int) bool{
  var (
    remote = host + ":" + strconv.Itoa(port)
  )
  tcpAddr, _ := net.ResolveTCPAddr("tcp4", remote)
  conn, err := net.DialTCP("tcp", nil, tcpAddr)
  if err != nil {
    return false
  }
  defer conn.Close()
  return true
}

func (s *Scan) echo(host string,port int){
	if s.html_==true{
		html:=s.get_title(host,port)
		pl(host+":"+strconv.Itoa(port)+" title:"+html)
	}else{
		pl(host+":"+strconv.Itoa(port))
	}
}

func (s *Scan) scanByPort(){
	var ip string
	for ;;{
		ip=s.next_ip()
		if ip==""{
			break
		}
		if s.connect(ip,s.port_){
			s.echo(ip,s.port_)
		}
	}
	s.ch <- 1
}

func (s *Scan) next_ip() string{
	if len(s.ip_list.m)==0{
		return ""
	}
	s.ip_list.Lock()
	ip:=s.ip_list.m[len(s.ip_list.m)-1]
	delete(s.ip_list.m,len(s.ip_list.m)-1)
	s.ip_list.Unlock()
	return ip
}


func (s *Scan) next_port() int{
	if len(s.port_list.m)==0{
		return 0
	}
	s.port_list.Lock()
	port:=s.port_list.m[len(s.port_list.m)-1]
	delete(s.port_list.m,len(s.port_list.m)-1)
	s.port_list.Unlock()
	return port
}

func (s *Scan) scanByIp(){
	var port int
	for ;;{
		port=s.next_port()
		if port==0{
			break
		}
		if s.connect(s.ip_,port){
			s.echo(s.ip_,port)
		}
	}
	s.ch <- 1
}

func (s *Scan) run(){
	if s.model==1{
		if s.connect(s.ip_,s.port_){
			s.echo(s.ip_,s.port_)
		}
	}else if s.model==2{
		for i:=0;i<s.thread_num;i++{
			go s.scanByPort()
		}
	}else if s.model==3{
		for i:=0;i<s.thread_num;i++{
			go s.scanByIp()
		}
	}
	if s.thread_num>0{
		for i:=0;i<s.thread_num;i++{
			<- s.ch
		}
	}
}


func (s *Scan) Init(ip string,port string,html string,thread_num int) bool{
	s.ip_=ip
	s.thread_num=thread_num
	var err error
	if port!=""{
		s.port_,err=strconv.Atoi(port)
		if err!=nil{
			pl("parse port error!")
			return false
		}
		if strings.Count(s.ip_,".")==3{
			s.thread_num=0
			s.model=1
		}else if (strings.Count(s.ip_,".")<3&&strings.Count(s.ip_,".")>0){
			s.ip_list.m=make(map[int]string)
			if strings.Count(s.ip_,".")==1{
				for i:=254;i>=0;i--{
					for j:=254;j>=0;j--{
						s.ip_list.m[len(s.ip_list.m)]=s.ip_+"."+strconv.Itoa(j)+"."+strconv.Itoa(i)
					}
				}
			}else{
				for i:=254;i>=0;i--{
					s.ip_list.m[len(s.ip_list.m)]=s.ip_+"."+strconv.Itoa(i)
				}
			}
			s.model=2
		}else{
			pl("error ip")
			return false
		}
	}else{
		if strings.Count(s.ip_,".")!=3{
			pl("when not port then ip must full")
			return false
		}
		s.port_=0
		s.model=3
		
		s.port_list.m=make(map[int]int)
		for i:=65534;i>0;i--{
			s.port_list.m[len(s.port_list.m)]=i
		}
	}
	if html=="0"{
		s.html_=false
	}else{
		s.html_=true
	}
	s.ch=make(chan int)
	return true
}

func main(){
		var ip_ =flag.String("ip","","input ip")
		var port_ =flag.String("p","","input port")
		var html_=flag.String("html","0","echo html title")
		var t_=flag.Int("t",200,"thread num")
		flag.Parse()
		scan:=new(Scan)
		if scan.Init(*ip_,*port_,*html_,*t_)==false{
			return
		}
		scan.run()
}
