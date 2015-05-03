# resolv4flx

Utility to resolve DNS entries in a file

- [Disclaimer](#disclaimer)
- [Usage](#usage)

## Disclaimer

This is a quick test app.

## Usage

Usage: resolv4flx [flags] ENTRY_FILE
	
flags:
	-w, -workers=5		Number of worker threads to resolve DNS entries

ENTRY_FILE:
	File containing "domain dnsType" entries, one per line


File entry example:

	thumbs2.ebaystatic.com.	AAAA
	s-static.ak.fbcdn.net.	A

Usage example:
	
	resolv4flx -w 10 query.txt

