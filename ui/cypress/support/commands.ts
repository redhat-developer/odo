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
    cy.get('[data-cy="'+value+'"]', { timeout: 60000 });
});

Cypress.Commands.add('selectTab', (n: number) => {
    cy.get('div[role=tab]').eq(n).click();
});

Cypress.Commands.add('setDevfile', (devfile: string) => {
    cy.get('[data-cy="yaml-input"]').type(devfile);
    cy.get('[data-cy="yaml-save"]').click();
});

Cypress.Commands.add('clearDevfile', () => {
    cy.get('[data-cy="yaml-clear"]', { timeout: 60000 }).click();
});

declare namespace Cypress {
    interface Chainable {
        getByDataCy(value: string): Chainable<void>
        selectTab(n: number):  Chainable<void>

        setDevfile(devfile: string): Chainable<void>
        clearDevfile(): Chainable<void>
    }
}
