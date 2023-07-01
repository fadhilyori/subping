# subping

subping is a powerful and user-friendly command-line tool that allows you to perform ICMP ping operations on all IP
addresses within a specified subnet range. With subping, you can effortlessly discover and monitor the availability of
devices within a network by systematically pinging each IP address within the defined subnet.

## Dependencies

Subping depends on the following third-party libraries:

- **go-ping**: https://github.com/go-ping/ping

## Documentation

The documentation for the Subping library can be found in the [docs](docs/) directory. It includes detailed information on how to use the library, examples, and API references.

The library consists of the following packages:

- **github.com/fadhilyori/subping**: The main package that provides the Subping struct and related functionalities.
- **github.com/fadhilyori/subping/pkg/network**: A subpackage that offers network-related utilities for working with IP addresses and subnet ranges.

Please refer to the documentation for the respective packages to understand how to use them in your applications.

## Usage

To use subping, follow these steps:

1. Install subping by downloading the latest release from
   the [releases page](https://github.com/fadhilyori/subping/releases).

2. Open a terminal or command prompt and navigate to the directory where subping is installed.

3. Run the subping command with the specified subnet range:

   ```shell
   subping [OPTIONS] <network subnet>
   ```

   **Options:**

   `-c <count>`: Specifies the number of ping attempts for each IP address. The default value is 3.

   `-n <numJobs>`: Specifies the number of maximum concurrent jobs spawned to perform ping operations. The default value
   is equal to the number of CPUs available on the system.

   `-t <timeout>`: Specifies the maximum ping timeout duration. The default value is "300ms". The timeout can be
   expressed in various units

## Import as Go Package

To use the Subping library, follow these steps:

1. Import the Subping package:

```go
import (
    "github.com/your-username/subping"
    "github.com/your-username/subping/pkg/network"
)
```

2. Create an instance of Subping by calling `NewSubping` with the desired options:

```go
subnetString := "172.17.0.0/24"
targets, err := network.GenerateIPListFromCIDRString(subnetString)
if err != nil {
    log.Fatal(err.Error())
}

opts := &subping.Options{
    Targets: targets,
    Count:   3,
    Timeout: 300 * time.Millisecond,
    NumJobs: 8,
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

4. Retrieve the results using the `GetResults` method:

```go
results := sp.GetResults()
```

The `results` variable will contain a map where the keys are the IP addresses, and the values are `*ping.Statistics`
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

## Examples

Here are a few examples of how to use subping:

Ping all IP addresses in the subnet range 172.17.0.0/24:

```shell
subping -t 300ms -c 3 -n 100 172.17.0.0/24
```

![](assets/images/usage-example.png?raw=true)

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a
pull request. For more details, see our [contribution guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).
