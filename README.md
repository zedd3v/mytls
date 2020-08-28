# myTls

Mimic TLS/JA3 fingerprint inside Node with help from Go

## Installation

```bash
$ npm install mytls
```

## Usage

```javascript
const initMyTls = require('mytls');

// Typescript: import initMyTls from 'mytls';

(async () => {
  const myTls = await initMyTls();

  const response = await myTls('https://ja3er.com/json', {
    body: '',
    headers: {
      'user-agent': 'customheaders',
    },
    ja3: '771,255-49195-49199-49196-49200-49171-49172-156-157-47-53,0-10-11-13,23-24,0',
    proxy: 'http://user@pass:ip:port'
  });
})();

// => {
// =>   status: 200,
// =>   body: '{"ja3_hash":"6fa3244afc6bb6f9fad207b6b52af26b", "ja3": "771,255-49195-49199-49196-49200-49171-49172-156-157-47-53,0-10-11-13,23-24,0", "User-Agent": "customheaders"}',
// =>   headers: {
// =>     'Access-Control-Allow-Origin': '*',
// =>     Connection: 'keep-alive',
// =>     'Content-Length': '285',
// =>     'Content-Type': 'application/json',
// =>     Date: 'Sat, 25 Jul 2020 20:43:43 GMT',
// =>     Server: 'nginx',
// =>     'Set-Cookie': [ 'visited=6fa3244agc6xx6f9fad007b6b52af26b' ]
// => }
```

## Maintainer

[![ZedDev](https://github.com/zedd3v.png?size=100)](https://abck.dev/)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)

[JA3Transport](https://github.com/cucyber/JA3Transport) & [uTls](https://github.com/refraction-networking/utls) used for the go service
