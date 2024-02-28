package main

import (
    "context"
    "os"
    "net/netip"
    
    "github.com/nornir-automation/gornir/pkg/gornir"
    "github.com/nornir-automation/gornir/pkg/plugins/connection"
    "github.com/nornir-automation/gornir/pkg/plugins/inventory"
    "github.com/nornir-automation/gornir/pkg/plugins/logger"
    "github.com/nornir-automation/gornir/pkg/plugins/output"
    "github.com/nornir-automation/gornir/pkg/plugins/runner"
    "github.com/nornir-automation/gornir/pkg/plugins/task"
)

type runiperf struct {
}

func (t *runiperf) Metadata() *gornir.TaskMetadata {
    return nil
}

func (t *runiperf) Run(ctx context.Context, logger gornir.Logger, host *gornir.Host) (gornir.TaskInstanceResult, error) {
    resOpen, err := (&connection.SSHOpen{}).Run(ctx, logger, host)
    if err != nil {
        return resOpen, err
    }

    // Server IP is client IP-1
    client_ip,_ := netip.ParseAddr(host.Hostname)
    server_ip := client_ip.Prev()
    res, err := (&task.RemoteCommand{Command: "iperf3 -c "+server_ip.String()}).Run(ctx, logger, host)
    if err != nil {
        return res, err
    }

    resClose, err := (&connection.SSHClose{}).Run(ctx, logger, host)
    if err != nil {
        return resClose, err
    }

    return res, nil
}

func main() {
    log := logger.NewLogrus(false)

    file := "hosts.yaml"
    plugin := inventory.FromYAML{HostsFile: file}
    inv, err := plugin.Create()
    if err != nil {
        log.Fatal(err)
    }

    server_filter := func(h *gornir.Host) bool {
        return h.Platform == "server"
    }
    client_filter := func(h *gornir.Host) bool {
        return h.Platform == "client"
    }

    rnr := runner.Parallel()
    gr := gornir.New().WithInventory(inv).WithLogger(log).WithRunner(rnr)

    // Open an SSH connection towards the devices
    _, err = gr.RunSync(
        context.Background(),
        &connection.SSHOpen{},
    )
    if err != nil {
        log.Fatal(err)
    }

    // defer closing the SSH connection we just opened
    defer func() {
        _, err = gr.RunSync(
            context.Background(),
            &connection.SSHClose{},
        )
        if err != nil {
            log.Fatal(err)
        }
    }()

    serverGroup := gr.Filter(server_filter)
    clientGroup := gr.Filter(client_filter)

    // Start iperf server on servers
    //res := make(chan *gornir.JobResult, len(gr.Inventory.Hosts))
    _, err = serverGroup.RunSync(
        context.Background(),
        &task.RemoteCommand{Command: "iperf3 -s -D"},
    )
    if err != nil {
        log.Fatal(err)
    }
    // next call is going to print the result on screen
    //output.RenderResults(os.Stdout, res, "Iperf servers", true)

    // Start iperf clients
    log.Info("Starting clients")
    results, err := clientGroup.RunSync(
        context.Background(),
        &runiperf{},
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Info("Done clients")
    output.RenderResults(os.Stdout, results, "Iperf clients", true)
}
