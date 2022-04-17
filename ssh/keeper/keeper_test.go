package keeper_test

import (
	"fmt"

	"github.com/whoisnian/glb/ssh/keeper"
)

const (
	sshAddr    string = "127.0.0.1:22"
	sshUser    string = "nian"
	sshKeyFile string = "~/.ssh/id_rsa"
	cmd        string = "uname -a"
)

func ExampleNewKeeper() {
	k := keeper.NewKeeper()
	if err := k.PreparePrivateKey(sshKeyFile); err != nil {
		panic(err)
	}

	client, err := k.NewClient(sshAddr, sshUser, sshKeyFile)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	out, err := client.Run(cmd)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// Linux alegion 5.12.14-arch1-1 #1 SMP PREEMPT Thu, 01 Jul 2021 07:26:06 +0000 x86_64 GNU/Linux
}
