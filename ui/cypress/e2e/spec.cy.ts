import {TAB_YAML, TAB_COMMANDS, TAB_CONTAINERS, TAB_IMAGES, TAB_METADATA, TAB_RESOURCES, TAB_EVENTS, TAB_VOLUMES} from './consts';

describe('devfile editor spec', () => {

  let originalDevfile: string
  before(() => {
    cy.readFile('devfile.yaml', null).then(yaml => originalDevfile = (<BufferType> yaml).toString())
  })

  afterEach(() => {
    cy.readFile('devfile.yaml', null).then(yaml => {
      if (originalDevfile !== (<BufferType> yaml).toString()) {
        cy.writeDevfileFile(originalDevfile)
      }
    });
  })

  it('displays matadata.name set in YAML', () => {
    cy.init();
    cy.fixture('input/with-metadata-name.yaml').then(yaml => {
      cy.setDevfile(yaml);
    });

    cy.selectTab(TAB_METADATA);
    cy.getByDataCy("metadata-name").should('have.value', 'test-devfile');
  });

  it('displays container set in YAML', () => {
    cy.init();
    cy.fixture('input/with-container.yaml').then(yaml => {
      cy.setDevfile(yaml);
    });

    cy.selectTab(TAB_CONTAINERS);
    cy.getByDataCy('container-info').first()
      .should('contain.text', 'container1')
      .should('contain.text', 'nginx')
      .should('contain.text', 'the command to run')
      .should('contain.text', 'with arg');
  });

  it('displays a created container', () => {
    cy.init();

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-name').type('volume1');
    cy.getByDataCy('volume-size').type('512Mi');
    cy.getByDataCy('volume-ephemeral').click();
    cy.getByDataCy('volume-create').click();

    cy.selectTab(TAB_CONTAINERS);
    cy.getByDataCy('container-name').type('created-container');
    cy.getByDataCy('container-image').type('an-image');
    cy.getByDataCy('volume-mount-add').click();
    cy.getByDataCy('volume-mount-path').type("/mnt/vol1");
    cy.getByDataCy('volume-mount-name').click().get('mat-option').contains('volume1').click();
    cy.getByDataCy('container-create').click();

    cy.getByDataCy('container-info').first()
      .should('contain.text', 'created-container')
      .should('contain.text', 'an-image')
      .should('contain.text', 'volume1')
      .should('contain.text', '/mnt/vol1');
  });

  it('displays a created image', () => {
    cy.init();

    cy.selectTab(TAB_IMAGES);
    cy.getByDataCy('image-name').type('created-image');
    cy.getByDataCy('image-image-name').type('an-image-name');
    cy.getByDataCy('image-build-context').type('/path/to/build/context');
    cy.getByDataCy('image-dockerfile-uri').type('/path/to/dockerfile');
    cy.getByDataCy('image-create').click();

    cy.getByDataCy('image-info').first()
      .should('contain.text', 'created-image')
      .should('contain.text', 'an-image-name')
      .should('contain.text', '/path/to/build/context')
      .should('contain.text', '/path/to/dockerfile');
  });

  it('displays a created resource, with manifest', () => {
    cy.init();

    cy.selectTab(TAB_RESOURCES);
    cy.getByDataCy('resource-name').type('created-resource');
    cy.getByDataCy('resource-toggle-inlined').click();
    cy.getByDataCy('resource-manifest').type('a-resource-manifest');
    cy.getByDataCy('resource-create').click();

    cy.getByDataCy('resource-info').first()
      .should('contain.text', 'created-resource')
      .should('contain.text', 'a-resource-manifest');
  });

  it('displays a created resource, with uri (default)', () => {
    cy.init();

    cy.selectTab(TAB_RESOURCES);
    cy.getByDataCy('resource-name').type('created-resource');
    cy.getByDataCy('resource-uri').type('/my/manifest.yaml');
    cy.getByDataCy('resource-create').click();

    cy.getByDataCy('resource-info').first()
      .should('contain.text', 'created-resource')
      .should('contain.text', 'URI')
      .should('contain.text', '/my/manifest.yaml');
  });

  it('displays a created volume', () => {
    cy.init();

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-name').type('created-volume');
    cy.getByDataCy('volume-size').type('512Mi');
    cy.getByDataCy('volume-ephemeral').click();
    cy.getByDataCy('volume-create').click();

    cy.getByDataCy('volume-info').first()
      .should('contain.text', 'created-volume')
      .should('contain.text', '512Mi')
      .should('contain.text', 'Yes')
  });

  it('creates an exec command with a new container', () => {
    cy.init();

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-name').type('volume1');
    cy.getByDataCy('volume-size').type('512Mi');
    cy.getByDataCy('volume-ephemeral').click();
    cy.getByDataCy('volume-create').click();

    cy.selectTab(TAB_COMMANDS);
    cy.getByDataCy('add').click();
    cy.getByDataCy('new-command-exec').click();

    cy.getByDataCy('command-exec-name').type('created-command');
    cy.getByDataCy('command-exec-command-line').type('a-cmdline');
    cy.getByDataCy('command-exec-working-dir').type('/path/to/working/dir');
    cy.getByDataCy('select-container').click().get('mat-option').contains('(New Container)').click();
    cy.getByDataCy('container-name').type('a-created-container');
    cy.getByDataCy('container-image').type('an-image');
    cy.getByDataCy('volume-mount-add').click();
    cy.getByDataCy('volume-mount-path').type("/mnt/vol1");
    cy.getByDataCy('volume-mount-name').click().get('mat-option').contains('volume1').click();
    cy.getByDataCy('container-create').click();

    cy.getByDataCy('select-container').should('contain', 'a-created-container');
    cy.getByDataCy('command-exec-create').click();

    cy.getByDataCy('command-info').first()
      .should('contain.text', 'created-command')
      .should('contain.text', 'a-cmdline')
      .should('contain.text', '/path/to/working/dir')
      .should('contain.text', 'a-created-container');

    cy.selectTab(TAB_CONTAINERS);
    cy.getByDataCy('container-info').first()
      .should('contain.text', 'a-created-container')
      .should('contain.text', 'an-image')
      .should('contain.text', 'volume1')
      .should('contain.text', '/mnt/vol1');
  });

  it('creates an apply image command with a new image', () => {
    cy.init();

    cy.selectTab(TAB_COMMANDS);
    cy.getByDataCy('add').click();
    cy.getByDataCy('new-command-image').click();
    cy.getByDataCy('command-image-name').type('created-command');
    cy.getByDataCy('select-container').click().get('mat-option').contains('(New Image)').click();
    cy.getByDataCy('image-name').type('a-created-image');
    cy.getByDataCy('image-image-name').type('an-image-name');
    cy.getByDataCy('image-build-context').type('/context/dir');
    cy.getByDataCy('image-dockerfile-uri').type('/path/to/Dockerfile');
    cy.getByDataCy('image-create').click();

    cy.getByDataCy('select-container').should('contain', 'a-created-image');
    cy.getByDataCy('command-image-create').click();

    cy.getByDataCy('command-info').first()
      .should('contain.text', 'created-command')
      .should('contain.text', 'a-created-image');

    cy.selectTab(TAB_IMAGES);
    cy.getByDataCy('image-info').first()
      .should('contain.text', 'a-created-image')
      .should('contain.text', 'an-image-name')
      .should('contain.text', '/context/dir')
      .should('contain.text', '/path/to/Dockerfile');
  });

  it('creates an apply resource command with a new resource using manifest', () => {
    cy.init();

    cy.selectTab(TAB_COMMANDS);
    cy.getByDataCy('add').click();
    cy.getByDataCy('new-command-apply').click();
    cy.getByDataCy('command-apply-name').type('created-command');
    cy.getByDataCy('select-container').click().get('mat-option').contains('(New Resource)').click();
    cy.getByDataCy('resource-name').type('a-created-resource');
    cy.getByDataCy('resource-toggle-inlined').click();
    cy.getByDataCy('resource-manifest').type('spec: {}');
    cy.getByDataCy('resource-create').click();

    cy.getByDataCy('select-container').should('contain', 'a-created-resource');
    cy.getByDataCy('command-apply-create').click();

    cy.getByDataCy('command-info').first()
      .should('contain.text', 'created-command')
      .should('contain.text', 'a-created-resource');

    cy.selectTab(TAB_RESOURCES);
    cy.getByDataCy('resource-info').first()
      .should('contain.text', 'a-created-resource')
      .should('contain.text', 'spec: {}');
  });

  it('creates an apply resource command with a new resource using uri (default)', () => {
    cy.init();

    cy.selectTab(TAB_COMMANDS);
    cy.getByDataCy('add').click();
    cy.getByDataCy('new-command-apply').click();
    cy.getByDataCy('command-apply-name').type('created-command');
    cy.getByDataCy('select-container').click().get('mat-option').contains('(New Resource)').click();
    cy.getByDataCy('resource-name').type('a-created-resource');
    cy.getByDataCy('resource-uri').type('/my/manifest.yaml');
    cy.getByDataCy('resource-create').click();

    cy.getByDataCy('select-container').should('contain', 'a-created-resource');
    cy.getByDataCy('command-apply-create').click();

    cy.getByDataCy('command-info').first()
      .should('contain.text', 'created-command')
      .should('contain.text', 'a-created-resource');

    cy.selectTab(TAB_RESOURCES);
    cy.getByDataCy('resource-info').first()
      .should('contain.text', 'a-created-resource')
      .should('contain.text', 'URI')
      .should('contain.text', '/my/manifest.yaml');
  });

  it('reloads the Devfile upon changes in the filesystem', () => {
    cy.init();
    cy.fixture('input/devfile-new-version.yaml').then(yaml => {
      cy.writeDevfileFile(yaml);
    });

    cy.selectTab(TAB_METADATA);
    cy.getByDataCy("metadata-name").should('have.value', 'my-component');

    cy.selectTab(TAB_CONTAINERS);
    cy.getByDataCy('container-info').first()
        .should('contain.text', 'my-cont1')
        .should('contain.text', 'some-image:latest')
        .should('contain.text', 'some command')
        .should('contain.text', 'some arg');
  });

  it('adds an event with an existing command', () => {
    cy.init();
    cy.fixture('input/with-exec-command.yaml').then(yaml => {
      cy.setDevfile(yaml);
    });
    cy.selectTab(TAB_EVENTS);
    cy.get('[data-cy="prestop"] [data-cy="input"]').click().type("{downArrow}{enter}");
    cy.selectTab(TAB_YAML);
    cy.get('[data-cy="yaml-input"]').should("contain.value", "events:\n  preStop:\n  - command1");
    cy.selectTab(TAB_EVENTS);
    cy.get('[data-cy="prestop"] button.mat-mdc-chip-remove').click();
    cy.selectTab(TAB_YAML);
    cy.get('[data-cy="yaml-input"]').should("contain.value", "events: {}");
  });
});
