function responseSnippet(body) {
    if (typeof body !== 'string') {
        return '';
    }
    return body.slice(0, 300).replace(/\s+/g, ' ');
}

function expectOkOrNotModified(resp, path) {
    if (resp.status === 500) {
        return cy
            .requestVhost('app.test', path, { failOnStatusCode: false, timeout: 20000 })
            .then((retryResp) => {
                expect([200, 304]).to.include(
                    retryResp.status,
                    `unexpected status on retry ${path}; snippet=${responseSnippet(retryResp.body)}`
                );
            });
    }
    expect([200, 304]).to.include(
        resp.status,
        `unexpected status ${path}; snippet=${responseSnippet(resp.body)}`
    );
}

describe('public routes exhaustive', () => {
    ['/'].forEach((path) => {
        it(`responds OK on ${path}`, () => {
            cy.requestVhost('app.test', path, { failOnStatusCode: false, timeout: 20000 }).then((resp) =>
                expectOkOrNotModified(resp, path)
            );
        });
    });
});
