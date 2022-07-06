# Ohm

![Go Report Card](https://goreportcard.com/badge/github.com/pnovotnak/ohm)

Ohm is a minimal app that helps break the cycle of endless scrolling. Ohm works by blocking domain fragments using 
[NextDNS](https://nextdns.com/) _Denylist_ feature. It works best with websites but may also work with some apps.

Ohm gives you allowances which can be configured by domain. Each allowance can be configured with a session duration 
and a lockout period once the allowance is exhausted. An optional "cooldown" period can also be provided.

If `cooldown` is not provided, Ohm is activated for a domain when you first query the domain (usually visiting the 
website). Once it's activated, it starts a timer for that domain. Once the timer expires, it enables the corresponding 
Denylist entry.

If `cooldown` is provided, Ohm will monitor DNS query logs for the domain. If no requests are made for the time specified 
by `cooldown`, the block will not be inserted.

_Note that browsers maintain very long-lived connections and typically systems have at least 2 layers of DNS caching. 
You may need to spend some time configuring Ohm for your system, or your system for Ohm._

# Configuration

See [`example-config.yaml`](example-config.yaml) for an example of how to configure Ohm.

# Deployment

The easiest way to deploy Ohm is with [Fly](https://fly.io/).

## Prerequisites

1. A [NextDNS](https://nextdns.com/) account (free).
   1. Add a NextDNS configuration profile for yourself.
      1. Enable query logging (`Settings` -> `Logs`). Ohm uses these logs to function.
         1. You may disable `Log clients IPs` if you like, and retention can be dropped to 1h.
      2. Create a `Denylist` entry for each site you wish to block.
      3. Note your profile's ID. This can be found in the URL when configuring your profile or from the `Setup` page.
      4. Retrieve your API token from your `Account` page.
2. A [Fly.io](https://fly.io/) account (also free).
   1. Configure any device (computer, phone etc.) That you'd like to use Ohm with to use the NextDNS profile created above.

## Setup & Deployment

1. Clone this repository.
2. Create a configuration file in `cmd/ohm/config.yaml`. See [Configuration](#configuration) for details.
3. Deploy Ohm to Fly
   ```shell
   cd ohm/
   flyctl deploy
   flyctl secrets set OHM_NEXTDNS_KEY=<API token from above> OHM_NEXTDNS_PROFILE=<your profile ID from above>
   # For good measure...
   flyctl restart
   ```
4. Monitor logs.
   ```shell
   flyctl logs
   ```
5. Try visiting sites you've configured Ohm to monitor.
