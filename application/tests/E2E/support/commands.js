Cypress.Commands.add('visitVhost', (host, path = '/') => {
    cy.visit(`https://127.0.0.1${path}`, {
        headers: { Host: host },
        failOnStatusCode: false,
    });
});

Cypress.Commands.add('requestVhost', (host, path = '/', options = {}) => {
    return cy.request({
        url: `https://127.0.0.1${path}`,
        headers: { Host: host },
        failOnStatusCode: false,
        ...options,
    });
});
