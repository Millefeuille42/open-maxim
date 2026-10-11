# Patches

This directory contains [Pnach-formatted patch
files](https://pcsx2.net/docs/advanced/writing-patches/) that are compatible
with the PCSX2 emulator.

The patch file needs to be named `SLUS-20090_718C85F0.pnach`. It needs to be
placed in the `patches/` directory in your PCSX2 data directory. To open your
data directory, use the `Tools -> Open Data Directory...` menu item within the
main emulator window.

## List of patches

### Bypass MAC address check

This patch is mandatory for networking to work properly on the PCSX2 emulator.

There are two versions of this patch:

- Fast boot (recommended): use this version if you're using the "Fast Boot"
  option to skip the PS2 startup screen
- No fast boot: use this version if you're _not_ using the "Fast Boot" option -
  you also need to select the patch that corresponds to your specific BIOS
  region/version

During startup, MaxIM's DEV9 IRX gets the MAC address of the network adapter
and its checksum that are stored in the adapter EEPROM. It then computes the
checksum from the MAC address that it read, and checks what the computed
checksum with the stored checksum. If this check fails, the result is that no
packet will be sent over the wire, even though MaxIM still looks to have
started up correctly and doesn't show any error to the user.

At the time of writing, PCSX2 emulates that the first read to the EEPROM
doesn't actually return the first bit at the requested address, but rather some
value that indicates whether there was an error or not. This matches behavior
observed in Sony's official IRX. 

However, MaxIM's IRX actually expects the first read to the EEPROM to return
the first bit of the MAC address. This leads to the MAC address and checksum
read by this IRX to be shifted by one bit. Due to the simplicity of the
checksum algorithm, the check sometimes passes, but oftentimes just silently
fails.

This fixes this problem by replacing the instruction that does this check with
a NOP instruction.

It is currently unknown whether this is due to an emulation inaccuracy, or if
MaxIM's IRX implementation really is broken.
