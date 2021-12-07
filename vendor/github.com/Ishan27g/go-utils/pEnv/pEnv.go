
package pEnv

import (
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/joho/godotenv"
)
type process struct {
    lock sync.Mutex
    clearEnvDelay time.Duration
}
// Init pEnv with the time taken by processes to read env.
// Each process is launched in a go routine. After processStartupTime duration, the env will be cleared
//
// Eg.
// func main(){
//    pEnv := pEnv.Init(time.Millisecond * 3)
//    pEnv.CmdFromFile(".env", "./main", nil)
//    pEnv.CmdFromMap(map[string]string{"HTTP_PORT": "7001", "HOST": "localhost"}, "./main", nil)
//    <- make(chan bool)
// }
func Init(processStartupTime time.Duration) EnvP {

    p := process{clearEnvDelay: processStartupTime, lock: sync.Mutex{}}
    return &p
}
type EnvP interface {
    CmdFromMap(envMap map[string]string, cmd string, cmdArgs []string)
    CmdFromFile(envFile string, cmd string, cmdArgs []string)
}

func (p *process) CmdFromMap(envMap map[string]string, cmd string, cmdArgs []string) {
    for k,v := range envMap{
        os.Setenv(k,v)
    }
    p.run(cmd, cmdArgs)
}

func (p *process) CmdFromFile(envFile string, cmd string, cmdArgs []string) {
    err := godotenv.Load(envFile)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
    p.run(cmd, cmdArgs)
}

func (p *process)run(cmd string, cmdArgs []string) {
    defer os.Clearenv()
    p.lock.Lock()
    go func() {
        err := godotenv.Exec(nil, cmd, cmdArgs)
        if err != nil {
            fmt.Println(err.Error())
            os.Exit(1)
        }
    }()
    <- time.After(p.clearEnvDelay) // wait for env to be read before clearing
    p.lock.Unlock()
}


