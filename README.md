# Currency Conversion Tool

Simple CLI tool to convert currency

## Install

Run 

```
go get github.com/ilyakaznacheev/fiatconv
``` 

and then call `fiatconv -h` to get usage instructions

## Usage

The tool API is 

```
fiatconv [OPTIONS] [AMOUNT] [SRC] [DST]
```

Where

Options:

- `--api-url`: Exchange API address (optional)
- `--proxy`: Proxy server path (optional)

Positional parameters:

- `AMOUNT`: decimal amount of source currency;
- `SRC`: ISO currency code of source currency;
- `DST`: ISO currency code of destination currency;

## Example

```bash
> fiatconv 1 USD RUB
USD 1.00 -> RUB 64.98

> fiatconv 1 USD JPY
USD 1.00 -> JPY 119
```