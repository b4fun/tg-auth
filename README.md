<h3 align="center">
<a href="https://github.com/b4fun/tg-auth">
<img src="docs/assets/logo.svg" width="180
px" heigh="auto" style="inline-block" />
</a>
</h3>

<h4 align="center">
Authentication/authorization middleware with Telegram.
</h4>

**tg-auth** enables drop-in [Telegram based authn/z][telegram-login] to your service. No modifications needed for your application's code.

[telegram-login]: https://core.telegram.org/widgets/login

![](./docs/assets/diagram.png)

## Settings

tg-auth runs with following environment variables:

| Variable Name | Description | Sample |
|:-----------:|:---|:---|
| `BOT_NAME` | Telegram bot name | |
| `BOT_TOKEN` | Telegram bot token | |
| `SIGNIN_URL` | URL endpoint of the signin protoal | `https://example.com/signin` |
| `SIGNIN_REDIRECT_CALLBACK_URL` | URL endpoint of the Telegram signin callback | `https://example.com/signin/callback` |
| `SIGNIN_AFTER_SIGNIN_URL` | URL endpoint after succeeded authentication | `https://example.com` |
| `AUTHZ_CHANNEL_IDS` | `,` separated Telegram channel id for checking user access permission | `-12345,-67890` |
| `AUTHZ_CACHE` | User access check cache TTL. Defaults to 5 minutes | `5m` |
| `AUTHN_COOKIE_SIGNING_KEY` | Base64 encoded AES signing key for encrypting cookie value. | |
| `AUTHN_COOKIE_NAME` | Name of the cookie | | 
| `AUTHN_COOKIE_DOMAIN` | The domain of the cookie | `example.com` |
| `AUTHN_SESSION_TTL` | TTL of the cookie. Defaults to 1 hour | `1h` |

## Examples

## LICENSE

MIT