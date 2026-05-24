const { defineConfig } = require('cypress');

module.exports = defineConfig({
    e2e: {
        specPattern: 'e2e/**/*.cy.js',
        supportFile: 'support/e2e.js',
        baseUrl: 'https://app.test',
        video: true,
        screenshotOnRunFailure: true,
        testIsolation: true,
        retries: { runMode: 1, openMode: 0 },
        setupNodeEvents(on) {
            on('task', {
                async dbQuery({ query, values = [] }) {
                    const pg = require('pg');
                    const client = new pg.Client({
                        host: 'postgres',
                        port: 5432,
                        user: 'app',
                        password: 'app',
                        database: 'app',
                    });
                    await client.connect();
                    try {
                        const result = await client.query(query, values);
                        return result.rows;
                    } finally {
                        await client.end();
                    }
                },
            });
        },
    },
    requestTimeout: 20000,
    responseTimeout: 20000,
});
