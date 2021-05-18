// ex: set tabstop=2:
package main
import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	//"strings"
	"time"
	"unicode"
	"unicode/utf8"
	. "unsafe"

	"github.com/Des-Nerger/go-astisub"
)
func bytesToString(b []byte) string {return *(*string)(Pointer(&b))}
func bytе(Bool bool) byte {return *(*byte)(Pointer(&Bool))}
func bооl(Byte byte) bool {return *(*bool)(Pointer(&Byte))}
func sign(i int) int {return i>>(strconv.IntSize-1)|int(bytе(i!=0))}
func panicIfNotNil(e error) {if e != nil {panic(e)}}
func abs(i int) int {m:=i>>(strconv.IntSize-1); return(i+m)^m}
func sqr(i int) int {return i*i}
const maxInt = 1<<(strconv.IntSize-1)-1

/*
func align(u, w []int) []int {
	fmt.Println(u, w)
	width:=len(w)-(len(u)-1)
	p:=make([]byte, (width*len(u)-1)/8+1)
	costs := make([]int, width+1); costs[width]=(maxInt/8)*7
	C := func(i, j int) *int {
		return &costs[func() int {
			if j==-1 || i==-1 && j!=0 {return width}
			return j
		} ()]
	}
	{	y,m:=-1,byte(0)
		for i := range u {
			for WSum,j:=0,0; j<width; j++ {
				W:=w[i+j]; WSum+=W
				fmt.Printf("%v+abs(%v-%v) %v+abs(%v-%v)\n", *C(i, j-1), u[i], WSum, *C(i-1, j), u[i], W)
				leftC, upC := *C(i, j-1)+abs(u[i]-WSum), *C(i-1, j)+abs(u[i]-W)
				if m==0 {y++; m=0b1000_0000}
				*C(-1, j), p[y] = func() (int, byte) { py:=p[y]
					if leftC <= upC {return leftC, py&^m}
					WSum=W; return upC, py|m
				} ()
				m>>=1
			}
			fmt.Println(costs[:len(costs)-1])
		}
	}
	s,I:=make([]int,len(u)),(len(u)-1)*width+(width-1); s[len(s)-1]=len(w)-1
	ym := func() (y int, m byte) {y=I/8; m=p[y]>>(7-I%8); return}
	for y,m:=ym();; {
		if I<0 {panic("unexpected I<0")}
		if m==0 {y--; m=p[y]}
		if m&1==1 {
			I-=width; y,m=ym()
			if I==-width {break}
			fmt.Println(s[:cap(s)])
			l:=len(s)-1; s[l-1]=s[l]-1; s=s[:l]
			continue
		}
		I--; m>>=1
		s[len(s)-1]--
	}
	s = s[:cap(s)]
	return s
}
*/

func cost(u,U []int) (c int) {for i,I:=0,max(len(u),len(U)); i<I; i++ {c+=abs(u[i]-U[i])}; return}
func ratioCost(u,U []int) (c float64) {
	for i,I:=0,max(len(u),len(U)); i<I; i++ {c+=math.Abs(1-float64(u[i])/float64(U[i]))}; return
}
func vectorCost(u,U []int) (c float64) {
	I:=max(len(u),len(U)); if I==1 {return 0}
	dp,su,sU := 0,0,0
	for i:=0; i<I; i++ {
		u, U := u[i], U[i]
		dp+=u*U; su+=sqr(u); sU+=sqr(U)
	}
	return 1-float64(dp)/math.Sqrt(float64(su*sU))
}

type alignment []int
func (a *alignment) g(i int) int {if i==-1 {return 0}; return (*a)[i]}
func (a *alignment) utterances(w []int, iup, iwp int) []int {
	u:=make([]int,len(*a))
	for i:=range u {
		for j,w:=range w[a.g(i-1):(*a)[i]] {
			ip := func()int{if w<0 {w*=-1; return iup}; return iwp}()
			u[i] += func()int{if j==0 {return 0}; return ip}() + w
		}
	}
	return u
}
func align(u, w []int, iup, iwp int) alignment {
	W:=len(w)-(len(u)-1); p:=make([]byte,W*len(u)); c,t:=make([]int,2*W),0
	for i,k:=0,0; i<len(u); func(){i++; t^=-1}() {
		N := func() int {switch i {case 0: return 1; case 1: k+=W-1}; return W} ()
		for n:=0; n<N; func(){k++; n++}() {
			C0, s := c[t&W+n], 0
			for j,J,k,P:=n,(t^-1)&W+n,k,byte(0); j<W; func(){j++; J++; k++; if P!=0xFF {P++}}() {
				w:=w[i+j]; ip:=func()int{if w<0 {w*=-1; return iup}; return iwp}()
				s += func()int{if j==n {return 0}; return ip}()+w
				C:=C0+abs(u[i]-s)
				if bооl(bytе(C<c[J]) | bytе(n==0)) {c[J], p[k] = C, P}
			}
		}
	}
	//t^=-1; for _, c := range [][]int{c[t&W+W-4 : t&W+W], c[(t^-1)&W+W-4 : (t^-1)&W+W]} {fmt.Println(c)}
	a := make(alignment, len(u))
	J,k := len(w), (len(u)-1)*W+(W-1)
	for i:=len(a)-1; i>=0; i-- {
		P:=int(p[k]); if P==0xFF {panic(fmt.Sprintf("wordCount >= %v", P+1))}
		a[i]=J; J-=P+1; k-=W+P
	}
	return a
}
/*...
	for i := range u {
		const prohibitC = (maxInt/8)*7
		for j,leftC,ws:=0,prohibitC,0; j<width; func(){j++;k++}() {
			W:=w[i+j]; ws+=W
			upC := func() int {if i==0 {if j==0 {return 0}; return prohibitC}; return c[j]} ()
			leftC, c[j], p[k] = func() (int, int, byte) {
				fmt.Printf("%v+abs(%v-%v)　%v+abs(%v-%v)　　", upC,u[i],W, leftC,u[i],ws)
				incedUpC, incedLeftC := upC+abs(u[i]-W), leftC+abs(u[i]-ws)
				if incedLeftC<=incedUpC {return leftC,incedLeftC,0}
				ws=W; return upC,incedUpC,1
			} ()
			fmt.Printf("%v,%v,%v\n", leftC, c[j], p[k])
		}
		fmt.Println(c)
	}
	a:=make([]int,len(u))
	J,k := func()(int,int){U,W:=len(u)-1,width-1; return U+W,U*width+W}()
	for k>=0 {
		a[len(a)-1]=J; J--
		if p[k]==0 {k--; continue}
		a=a[:len(a)-1]; k-=width
	}
	return a[:cap(a)]
*/

func min(a, b int) int {if a < b {return a}; return b}
func max(a, b int) int {if a > b {return a}; return b}
//func roundDiv(n, m int) int {return (n+m/2)/m}
func main() {
	d,m := [2]time.Duration{},min(3,len(os.Args)-2)
	for s,i:=os.Args[1:m],0; ; i++ {
		if i>=len(s) {if i==1 {d[i]=d[i-1]}; break}
		f,e:=strconv.ParseFloat(s[i],64); panicIfNotNil(e)
		d[i] = time.Duration(math.Round(f*float64(time.Second)))
	}
	fmt.Println(d)
	subs, e := astisub.OpenFile(os.Args[m]); panicIfNotNil(e)
	u, dSum := make([]int, 0, len(subs.Items)), 0
	for _, it := range subs.Items {
		u = append(u, int(((it.EndAt-d[1])-(it.StartAt+d[0])).Milliseconds()) )
		dSum += u[len(u)-1]
	}
	s := func() string {bytes,e:=ioutil.ReadAll(os.Stdin); panicIfNotNil(e); return bytesToString(bytes)} ()
	w, ws, totalLNCount := []int(nil), []int(nil), 0
	{	const signBit=-1<<(strconv.IntSize-1)
		var (
			lnStart,lnCount=-1,0
			a = func() {
				switch sign(lnCount) {
				case -1:
					lnCount = -(lnCount&^signBit)
					fallthrough
				case +1:
					w=append(w,lnCount); lnCount=0; ws=append(ws,lnStart); lnStart=-1
				}
			}
			pt=rune(utf8.RuneError); r rune; size int
		)
		for i, pr, s := 0, rune(utf8.RuneError), s; s!=""; func(){i+=size; pr=r; s=s[size:]}() {
			r, size = utf8.DecodeRuneInString(s)
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				const (q='“'; qlen=len(string(q)))
				if lnStart==-1 {
					lnStart = i - func()int{if pr==q {return qlen}; return 0}()
					switch pt {
					case '.','!','?',':','…': lnCount|=signBit
					}
				}
				lnCount++; totalLNCount++
				continue
			}; if lnStart!=-1 {pt=r}; a()
		}; a()
	}
	const iup,iwpWorth = 5*1000,1.5
	totalWorth := float64(totalLNCount) + float64(len(w)-len(u))*iwpWorth
	ratio := float64(dSum) / totalWorth
	iwp := int(math.Round(iwpWorth*ratio))
	for i := range w {w[i] = int(math.Round(float64(w[i])*ratio))}
	/**iup=iwp;/**/ a:=align(u,w,iup,iwp); fmt.Println(a)
	/**au:=a.utterances(w,iup,iwp)/**
	r:=alignment{4,7,13,18,23,28,35,39,43,48,53,57,62,66,69,73,77,80,84,89,93,97,104,111,114};ru:=r.utterances(w,iup,iwp)
	fmt.Println(ratio,"\n", cost(au,u),ratioCost(au,u),vectorCost(au,u),"\n", cost(ru,u),ratioCost(ru,u),vectorCost(ru,u))
	a=r/* / fmt.Printf("%.3f\n%v　%v　%v\n", ratio, cost(au,u),ratioCost(au,u),vectorCost(au,u))/**/
	for i := range a {
		subs.Items[i].Lines[0].Items[0].Text = s[ws[a.g(i-1)] : func() int {
			if a[i]==len(ws) {return len(s)}; return ws[a[i]]
		} ()]
	}
	e=subs.Write(os.Args[m+1]); panicIfNotNil(e)
}
