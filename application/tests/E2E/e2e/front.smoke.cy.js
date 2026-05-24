describe('frontend smoke', () => {
    it('loads homepage on app.test', () => {
        cy.visitVhost('app.test', '/');
        cy.contains(/hello world|environment/i).should('exist');
    });
});
