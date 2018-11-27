package main

import "iphash-daemon/entry"

func main() {
	entry.Start()
}

func main1() {
	// path, err := exec.LookPath("ping")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// proc, err := os.StartProcess(path, []string{"ping", "www.baidu.com"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// go (func() {
	// 	time.Sleep(time.Duration(1) * time.Second)
	// 	proc.Signal(os.Interrupt)
	// })()
	// proc.Wait()
	// fmt.Println("finish")
	// err := worker.DeCompress("test.tar.gz", "/tmp/")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// os.StartProcess("/bin/ping", []string{"ping", "-c 3", "www.baidu.com"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	// var (
	// 	stop = make(chan struct{})
	// 	done = make(chan struct{})
	// )
	// executor := &worker.Main{Done: done, Stop: stop}
	// executor.Start()

	// procCheck, err := os.StartProcess("/tmp/iphash/iphash-linux-amd64-v0.01/ipfs", []string{"ipfs", "stats", "bw"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	// if err != nil {
	// 	fmt.Printf("%#v", err)
	// }
	// procCheck.Wait()
	// fmt.Println("finish")

	// cmd := exec.Command("/tmp/iphash/iphash-linux-amd64-v0.01/ipfs", "stats", "bw")
	// var outb, errb bytes.Buffer
	// cmd.Stdout = &outb
	// cmd.Stderr = &errb
	// err := cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("out:", outb.String(), "err:", errb.String())
}
