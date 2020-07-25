import { spawn, ChildProcessWithoutNullStreams } from 'child_process';
import path from 'path';
import { EventEmitter } from 'events';
import { Server } from 'ws';

export interface MyTlsRequestOptions {
	headers?: {
		[key: string]: any;
	};
	body?: string;
	ja3?: string;
}

export interface MyTlsResponse {
	status: number;
	body: string;
	headers: {
		[key: string]: any;
	};
}

let child: ChildProcessWithoutNullStreams;

const cleanExit = (message?: string | Error) => {
	if (message) console.log(message);
	child.kill();
	process.exit();
};
process.on('SIGINT', () => cleanExit());
process.on('SIGTERM', () => cleanExit());

class Golang extends EventEmitter {
	server: Server;
	constructor(port: number) {
		super();

		child = spawn(path.join(__dirname, 'index.exe'), {
			env: { WS_PORT: port.toString() },
			shell: true,
			windowsHide: true,
		});

		child.stderr.on('data', (stderr) => {
			if (!stderr.toString().includes('REQUESTIDONTHELEFT')) {
				cleanExit(new Error('Invalid JA3 hash. Exiting... (Golang wrapper exception)'));
			}

			const splitRequestIdAndError = stderr.toString().split('REQUESTIDONTHELEFT');
			const [requestId, error] = splitRequestIdAndError;
			this.emit(requestId, { error });
		});

		this.server = new Server({ port });

		this.server.on('connection', (ws) => {
			this.emit('ready');

			ws.on('message', (data: string) => {
				const message = JSON.parse(data);
				this.emit(message.RequestID, message.Response);
			});
		});
	}

	request(
		requestId: string,
		options: {
			[key: string]: any;
		}
	) {
		[...this.server.clients][0].send(JSON.stringify({ requestId, options }));
	}
}

const initMyTls = (
	port: number = 9119
): Promise<{
	(
		url: string,
		opts: MyTlsRequestOptions,
		method?: 'head' | 'get' | 'post' | 'put' | 'delete' | 'trace' | 'options' | 'connect' | 'patch'
	): Promise<MyTlsResponse>;
	head(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	get(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	post(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	put(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	delete(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	trace(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	options(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	connect(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
	patch(url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse>;
}> => {
	return new Promise((resolveReady) => {
		const instance = new Golang(port);

		instance.on('ready', () => {
			const mytls = (() => {
				const MyTls = async (
					url: string,
					opts: MyTlsRequestOptions,
					method: 'head' | 'get' | 'post' | 'put' | 'delete' | 'trace' | 'options' | 'connect' | 'patch' = 'get'
				): Promise<MyTlsResponse> => {
					return new Promise((resolveRequest, rejectRequest) => {
						const requestId = `${url}${Math.floor(Date.now() * Math.random())}`;

						if (!opts.ja3) {
							opts.ja3 =
								'771,4865-4866-4867-49196-49195-49188-49187-49162-49161-52393-49200-49199-49192-49191-49172-49171-52392-157-156-61-60-53-47-49160-49170-10,65281-0-23-13-5-18-16-11-51-45-43-10-21,29-23-24-25,0';
						}

						if (!opts.body) {
							opts.body = '';
						}

						instance.request(requestId, {
							url,
							...opts,
							method,
						});

						instance.once(requestId, (response) => {
							if (response.error) rejectRequest(response.error);

							const { Status: status, Body: body, Headers: headers } = response;

							resolveRequest({
								status,
								body,
								headers,
							});
						});
					});
				};
				MyTls.head = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'head');
				};
				MyTls.get = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'get');
				};
				MyTls.post = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'post');
				};
				MyTls.put = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'put');
				};
				MyTls.delete = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'delete');
				};
				MyTls.trace = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'trace');
				};
				MyTls.options = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'options');
				};
				MyTls.connect = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'options');
				};
				MyTls.patch = (url: string, opts: MyTlsRequestOptions): Promise<MyTlsResponse> => {
					return MyTls(url, opts, 'patch');
				};

				return MyTls;
			})();
			resolveReady(mytls);
		});
	});
};

export default initMyTls;

// CommonJS support for default export
module.exports = initMyTls;
module.exports.default = initMyTls;
module.exports.__esModule = true;
