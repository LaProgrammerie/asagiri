import './commands';

afterEach(function () {
    if (this.currentTest.state !== 'failed') {
        return;
    }

    cy.url().then((url) => {
        Cypress.log({ name: 'failed-url', message: url });
    });

    cy.document().then((doc) => {
        const html = doc.documentElement.outerHTML.slice(0, 2000);
        Cypress.log({ name: 'failed-html-snippet', message: html });
    });
});
