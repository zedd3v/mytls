module.exports = {
	env: {
		es2020: true,
		amd: true,
		node: true,
	},
	extends: [
		'eslint:recommended',
		'plugin:@typescript-eslint/eslint-recommended',
		'plugin:@typescript-eslint/recommended',
	],
	parser: '@typescript-eslint/parser',
	parserOptions: {
		ecmaVersion: 11,
		sourceType: 'module',
	},
	plugins: ['@typescript-eslint', 'prefer-arrow'],
	rules: [
		{
			'prefer-arrow/prefer-arrow-functions': [
				'warn',
				{
					disallowPrototype: true,
					singleReturnOnly: false,
					classPropertiesAllowed: false,
				},
			],
		},
		{
			'prefer-arrow-callback': ['warn'],
		},
	],
};
