# open-maxim

This project is dedicated to reimplement the MaxIM Instant Messenger servers initially provided by Datel as their original servers are now permanently offline.

First objective is to (try to at least) provide 100% functionality on the server, and later expand it and
integrate it to other chat platforms and protocols like IRC and Matrix.

The project is for now in its early phase, as I am reverse engineering the original binary and trying to map out the protocol.

Complementary efforts, like a complete decomp of the original PS2 binary can be integrated in the scope of this projet.

All my research, packet captures, disassemblies, scripts, notes, etc. are committed directly to this repository when relevant.

## Connecting to a server

The server has been tested and developed using PCSX2 as a reference
other emulator supporting networking (DEV9) and official hardware with a custom firmware
that allows to edit DNS records should work too.

To access, modify your DNS to point `www.datelversions.com` to the target server, 
the official open-maxim server is at `141.145.193.112`.

Using a USB keyboard is recommended, using the controller works but is tedious.

You can now register, log in and chat, the forum section is yet to be implemented.

### PCSX2

On PCSX2, you will need to enable one of the "Bypass MAC address check" patches
to make sure networking works properly. Please check the
[`patches/`](./patches/) directory for more information.

In the "Network & HDD" settings of the emulator, configure the options as
follows:

- Enabled: checked
- Ethernet Device Type: Sockets
- Ethernet Device: Auto
- Intercept DHCP tab:
    - DNS1 Address: select Internal
    - DNS2 Address: select Auto
- Internal DNS tab, add a row with the following:
    - Name: Datel
    - Url: `www.datelversions.com`
    - Address: `141.145.193.112`, or the IP address of your custom server
    - Enabled: checked

When testing, we'd be happy if you provide us the hardware, emulator, BIOS and
MaxIM version you are using.

## Contributing

Contributions are open and welcome. You can help by:

- Reverse Engineering of the original PS2 client.
- Documentation of the chat protocol.
- Server / tooling development.
- Alternative client development.
- Providing network captures (`.pcap`), legit retail disc dumps, hardware testing, additional, lore clients etc.

### How to Submit

- Open an issue or PR for spec proposals, code contributions, etc.
- Code contributions sent out by email are also accepted.
- For informal questions or sharing research notes, open a GitHub Discussion or reach out directly with the contact info available on my profile.

To maintain clean lineage and clear copyright tracking, all committed code must include a `Signed-off-by` line 
(using `git commit -s` to comply with the [Developer Certificate of Origin](https://developercertificate.org/)) 
and, ideally, be GPG/SSH signed.

All of my research, scripts, notes are made available in this repository. Feel free to use them as reference, tools, etc. for your contributions.

### AI Policy

For AI usage, please refer to [Matrix's AI policy](https://github.com/matrix-org/matrix.org/blob/main/CONTRIBUTING.md#ai-policy) 

## License

This project is licensed under the GNU Affero General Public License v3.0.
