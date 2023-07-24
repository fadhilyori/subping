# subping

subping is a powerful and user-friendly command-line tool that allows you to perform ICMP ping operations on all IP
addresses within a specified subnet range. With subping, you can effortlessly discover and monitor the availability of
devices within a network by systematically pinging each IP address within the defined subnet.

## Dependencies

Subping depends on the following third-party libraries:

- **pro-bing** : https://github.com/prometheus-community/pro-bing
- **go-figure** : https://github.com/common-nighthawk/go-figure
- **cobra** : https://github.com/spf13/cobra
- **logrush** : https://github.com/sirupsen/logrus
- **network** : https://github.com/fadhilyori/subping/pkg/network

## Documentation

The documentation for the Subping library can be found at https://pkg.go.dev/github.com/fadhilyori/subping. It includes detailed information on how to use the library, and examples.

The library consists of the following packages:

- **[github.com/fadhilyori/subping](https://pkg.go.dev/github.com/fadhilyori/subping)**: The main package that provides the Subping struct and related functionalities.
- **[github.com/fadhilyori/subping/pkg/network](https://pkg.go.dev/github.com/fadhilyori/subping/pkg/network)**: A subpackage that offers network-related utilities for working with IP addresses and subnet ranges.

Please refer to the documentation for the respective packages to understand how to use them in your applications.

## Usage

To use subping, follow these steps:

1. Install subping by downloading the latest release from
   the [releases page](https://github.com/fadhilyori/subping/releases).

2. Open a terminal or command prompt and navigate to the directory where subping is installed.

3. Run the subping command with the specified subnet range:

   ```shell
   subping [flags] [network subnet]
   ```

The following flags are available for the `subping` command:

- `-c, --count int`: Specifies the number of ping attempts for each IP address. (default 1)
- `-h, --help`: Displays help information for the `subping` command.
- `-i, --interval string`: Specifies the time duration between each ping request. (default "300ms")
- `-n, --job int`: Specifies the number of maximum concurrent jobs spawned to perform ping operations. (default 128)
- `--offline`: Specify whether to display the list of offline hosts.
- `-t, --timeout string`: Specifies the maximum ping timeout duration for each ping request. (default "80ms")
- `-v, --version`: Displays the version information for `subping`.

## Examples

Here are a few examples of how to use subping:

Ping all IP addresses in the subnet range 172.17.0.0/24:

```shell
subping -t 300ms -c 3 -n 100 172.17.0.0/24
```

![](assets/images/usage-example.png?raw=true)

## Import as Go Package

To use the Subping library, follow these steps:

1. Import the Subping package:

    ```go
    import (
        "github.com/fadhilyori/subping"
    )
    ```

2. Create an instance of Subping by calling `NewSubping` with the desired options:

    ```go
    opts := &subping.Options{
        LogLevel: "debug",
        Subnet: "172.17.0.0/24",
        Count:   3,
        Interval: 1 * time.Second,
        Timeout: 3 * time.Second,
        MaxWorkers: 8,
    }
    
    sp, err := subping.NewSubping(opts)
    if err != nil {
        log.Fatal(err)
    }
    
    ```

Note: Ensure that you have imported the necessary packages, such as `"time"` and `"log"`.

3. Run the Subping process by calling the `Run` method:

    ```go
    sp.Run()
    ```

This will initiate the ICMP ping operations on the specified IP addresses.

4. Retrieve the results:

    ```go
    results := sp.Results
    ```

    The `results` variable will contain a map where the keys are the IP addresses, and the values are `*subping.Result`
    representing the ping statistics for each IP address.

5. Optionally, you can use the `GetOnlineHosts` method to filter the results and obtain only the IP addresses that
   responded
   to the ping:

    ```go
    onlineHosts := sp.GetOnlineHosts()
    ```

    The `onlineHosts` variable will contain a map of the online IP addresses and their corresponding ping statistics.

6. You can also call the `RunPing` function directly to perform a ping operation on a single IP address:

    ```go
    ipAddress := net.ParseIP("192.168.1.1")
    count := 3
    timeout := 300 * time.Millisecond
    
    stats := subping.RunPing(ipAddress, count, timeout)
    ```

    The `stats` variable will contain the ping statistics for the specified IP address.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a
pull request. For more details, see our [contribution guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).
