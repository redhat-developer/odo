/// <reference types="cypress" />
// ***********************************************
// This example commands.ts shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })
//
// declare global {
//   namespace Cypress {
//     interface Chainable {
//       login(email: string, password: string): Chainable<void>
//       drag(subject: string, options?: Partial<TypeOptions>): Chainable<Element>
//       dismiss(subject: string, options?: Partial<TypeOptions>): Chainable<Element>
//       visit(originalFn: CommandOriginalFn, url: string, options: Partial<VisitOptions>): Chainable<Element>
//     }
//   }
// }

Cypress.Commands.add('getByDataCy', (value: string) => {
    cy.get('[data-cy="'+value+'"]');
});

Cypress.Commands.add('selectTab', (n: number) => {
    cy.get('div[role=tab]').eq(n).click();
});

Cypress.Commands.add('init', () => {
    cy.intercept('GET', '/api/v1/devfile').as('init.fetchDevfile');
    cy.intercept('PUT', '/api/v1/devstate/devfile').as('init.applyDevState');
    cy.visit('http://localhost:4200');
    cy.wait(['@init.fetchDevfile', '@init.applyDevState']);

    cy.clearDevfile()
});

Cypress.Commands.add('setDevfile', (devfile: string) => {
    cy.intercept('PUT', '/api/v1/devstate/devfile').as('setDevfile.applyDevState');
    cy.get('[data-cy="yaml-input"]').type(devfile);
    cy.get('[data-cy="yaml-save"]').click();
    cy.wait(['@setDevfile.applyDevState']);
});

Cypress.Commands.add('clearDevfile', () => {
    cy.intercept('DELETE', '/api/v1/devstate/devfile').as('clearDevfile.clearDevState');
    cy.intercept('PUT', '/api/v1/devstate/devfile').as('clearDevfile.applyDevState');
    cy.get('[data-cy="yaml-clear"]', { timeout: 60000 }).click();
    cy.wait(['@clearDevfile.clearDevState', '@clearDevfile.applyDevState']);
});

// writeDevfileFile writes the specified content into the local devfile.yaml file on the filesystem.
// Since #6902, doing so sends notification from the server to the client, and makes it reload the Devfile.
Cypress.Commands.add('writeDevfileFile', (content: string) => {
    cy.intercept('PUT', '/api/v1/devstate/devfile').as('writeDevfileFile.applyDevState');
    cy.writeFile('devfile.yaml',  content)
    cy.wait(['@writeDevfileFile.applyDevState']);
});

declare namespace Cypress {
    interface Chainable {
        init(): Chainable<void>

        getByDataCy(value: string): Chainable<void>
        selectTab(n: number):  Chainable<void>

        setDevfile(devfile: string): Chainable<void>
        clearDevfile(): Chainable<void>

        writeDevfileFile(content: string): Chainable<void>
    }
}
