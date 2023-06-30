# subping

subping is a powerful and user-friendly command-line tool that allows you to perform ICMP ping operations on all IP
addresses within a specified subnet range. With NetPing, you can effortlessly discover and monitor the availability of
devices within a network by systematically pinging each IP address within the defined subnet.

## Usage

To use subping, follow these steps:

1. Install subping by downloading the latest release from
   the [releases page](https://github.com/fadhilyori/subping/releases).

2. Open a terminal or command prompt and navigate to the directory where subping is installed.

3. Run the subping command with the specified subnet range:

   ```shell
   subping <subnet>
   ```

## Examples

Here are a few examples of how to use subping:

Ping all IP addresses in the subnet range 172.17.0.0/24:

```shell
subping 172.17.0.0/24
```

![](docs/images/usage-example.png?raw=true)

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a
pull request. For more details, see our [contribution guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).
