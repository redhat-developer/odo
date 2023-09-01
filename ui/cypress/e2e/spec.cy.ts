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

  it('displays a created container without source configuration', () => {
    cy.init();

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-name').type('volume1');
    cy.getByDataCy('volume-size').type('512Mi');
    cy.getByDataCy('volume-ephemeral').click();
    cy.getByDataCy('volume-create').click();

    cy.selectTab(TAB_CONTAINERS);
    cy.getByDataCy('container-name').type('created-container');
    cy.getByDataCy('container-image').type('an-image');
    cy.getByDataCy('container-env-add').click();
    cy.getByDataCy('container-env-name-0').type("VAR1");
    cy.getByDataCy('container-env-value-0').type("val1");
    cy.getByDataCy('container-env-plus').click();
    cy.getByDataCy('container-env-name-1').type("VAR2");
    cy.getByDataCy('container-env-value-1').type("val2");
    cy.getByDataCy('container-env-plus').click();
    cy.getByDataCy('container-env-name-2').type("VAR3");
    cy.getByDataCy('container-env-value-2').type("val3");

    cy.getByDataCy('volume-mount-add').click();
    cy.getByDataCy('volume-mount-path-0').type("/mnt/vol1", {force: true});
    cy.getByDataCy('volume-mount-name-0').click().get('mat-option').contains('volume1').click();

    cy.getByDataCy('endpoints-add').click();
    cy.getByDataCy('endpoint-name-0').type("ep1");
    cy.getByDataCy('endpoint-targetPort-0').type("4001");
    
    cy.getByDataCy('volume-mount-add').click();
    cy.getByDataCy('volume-mount-path-1').type("/mnt/vol2", {force: true});
    cy.getByDataCy('volume-mount-name-1').click().get('mat-option').contains('(New Volume)').click();
    cy.getByDataCy('volume-name').type('volume2');
    cy.getByDataCy('volume-create').click();

    cy.getByDataCy('container-more-params').click();
    cy.getByDataCy('container-deploy-anno-add').click();
    cy.getByDataCy('container-deploy-anno-name-0').type("DEPANNO1");
    cy.getByDataCy('container-deploy-anno-value-0').type("depval1");
    cy.getByDataCy('container-deploy-anno-plus').click();
    cy.getByDataCy('container-deploy-anno-name-1').type("DEPANNO2");
    cy.getByDataCy('container-deploy-anno-value-1').type("depval2");
    cy.getByDataCy('container-svc-anno-add').click();
    cy.getByDataCy('container-svc-anno-name-0').type("SVCANNO1");
    cy.getByDataCy('container-svc-anno-value-0').type("svcval1");
    cy.getByDataCy('container-svc-anno-plus').click();
    cy.getByDataCy('container-svc-anno-name-1').type("SVCANNO2");
    cy.getByDataCy('container-svc-anno-value-1').type("svcval2");

    cy.getByDataCy('container-create').click();

    cy.getByDataCy('container-info').first()
      .should('contain.text', 'created-container')
      .should('contain.text', 'an-image')
      .should('contain.text', 'VAR1: val1')
      .should('contain.text', 'VAR2: val2')
      .should('contain.text', 'VAR3: val3')
      .should('contain.text', 'volume1')
      .should('contain.text', '/mnt/vol1')
      .should('contain.text', 'volume2')
      .should('contain.text', '/mnt/vol2')
      .should('not.contain.text', 'Mount Sources')
      .should('contain.text', 'ep1')
      .should('contain.text', '4001')
      .should('contain.text', 'Deployment Annotations')
      .should('contain.text', 'DEPANNO1: depval1')
      .should('contain.text', 'DEPANNO2: depval2')
      .should('contain.text', 'Service Annotations')
      .should('contain.text', 'SVCANNO1: svcval1')
      .should('contain.text', 'SVCANNO2: svcval2');

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-info').eq(1)
    .should('contain.text', 'volume2');
  });

  it('displays a created container with source configuration', () => {
    cy.init();

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-name').type('volume1');
    cy.getByDataCy('volume-size').type('512Mi');
    cy.getByDataCy('volume-ephemeral').click();
    cy.getByDataCy('volume-create').click();

    cy.selectTab(TAB_CONTAINERS);
    cy.getByDataCy('container-name').type('created-container');
    cy.getByDataCy('container-image').type('an-image');
    cy.getByDataCy('container-more-params').click();
    cy.getByDataCy('container-sources-configuration').click();
    cy.getByDataCy('container-sources-specific-directory').click();
    cy.getByDataCy('container-source-mapping').type('/mnt/sources');

    cy.getByDataCy('volume-mount-add').click();
    cy.getByDataCy('volume-mount-path-0').type("/mnt/vol1", {force: true});
    cy.getByDataCy('volume-mount-name-0').click().get('mat-option').contains('volume1').click();

    cy.getByDataCy('volume-mount-add').click();
    cy.getByDataCy('volume-mount-path-1').type("/mnt/vol2", {force: true});
    cy.getByDataCy('volume-mount-name-1').click().get('mat-option').contains('(New Volume)').click();
    cy.getByDataCy('volume-name').type('volume2');
    cy.getByDataCy('volume-create').click();

    cy.getByDataCy('container-create').click();

    cy.getByDataCy('container-info').first()
      .should('contain.text', 'created-container')
      .should('contain.text', 'an-image')
      .should('contain.text', 'volume1')
      .should('contain.text', '/mnt/vol1')
      .should('contain.text', 'volume2')
      .should('contain.text', '/mnt/vol2')
      .should('contain.text', 'Mount Sources')
      .should('contain.text', '/mnt/sources');

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-info').eq(1)
    .should('contain.text', 'volume2');
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

    cy.getByDataCy('image-build-startup').first()
      .should('contain.text', 'Yes, the image is not referenced by any command');
  });

  it('displays a modified image', () => {
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

    cy.getByDataCy('image-build-startup').first()
      .should('contain.text', 'Yes, the image is not referenced by any command');

    cy.getByDataCy('image-edit').click();
    cy.getByDataCy('image-auto-build-always').click();
    cy.getByDataCy('image-image-name').type('{selectAll}{del}another-image-name');
    cy.getByDataCy('image-build-context').type('/new/path/to/build/context');
    cy.getByDataCy('image-dockerfile-uri').type('/new/path/to/dockerfile');
    cy.getByDataCy('image-save').click();
    
    cy.getByDataCy('image-info').first()
      .should('contain.text', 'created-image')
      .should('contain.text', 'another-image-name')
      .should('contain.text', '/new/path/to/build/context')
      .should('contain.text', '/new/path/to/dockerfile');

    cy.getByDataCy('image-build-startup').first()
      .should('contain.text', 'Yes, forced');

    });

  it('displays a created image with forced build', () => {
    cy.init();

    cy.selectTab(TAB_IMAGES);
    cy.getByDataCy('image-name').type('created-image');
    cy.getByDataCy('image-image-name').type('an-image-name');
    cy.getByDataCy('image-build-context').type('/path/to/build/context');
    cy.getByDataCy('image-dockerfile-uri').type('/path/to/dockerfile');
    cy.getByDataCy('image-auto-build-always').click();
    cy.getByDataCy('image-create').click();

    cy.getByDataCy('image-build-startup').first()
      .should('contain.text', 'Yes, forced');
  });

  it('displays a created image with disabled build', () => {
    cy.init();

    cy.selectTab(TAB_IMAGES);
    cy.getByDataCy('image-name').type('created-image');
    cy.getByDataCy('image-image-name').type('an-image-name');
    cy.getByDataCy('image-build-context').type('/path/to/build/context');
    cy.getByDataCy('image-dockerfile-uri').type('/path/to/dockerfile');
    cy.getByDataCy('image-auto-build-never').click();
    cy.getByDataCy('image-create').click();

    cy.getByDataCy('image-build-startup').first()
      .should('contain.text', 'No, disabled');
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

    cy.getByDataCy('resource-deploy-startup').first()
      .should('contain.text', 'Yes, the resource is not referenced by any command');
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

  it('displays a created resource, with forced deploy', () => {
    cy.init();

    cy.selectTab(TAB_RESOURCES);
    cy.getByDataCy('resource-name').type('created-resource');
    cy.getByDataCy('resource-toggle-inlined').click();
    cy.getByDataCy('resource-manifest').type('a-resource-manifest');
    cy.getByDataCy('resource-auto-deploy-always').click();
    cy.getByDataCy('resource-create').click();

    cy.getByDataCy('resource-info').first()
      .should('contain.text', 'created-resource')
      .should('contain.text', 'a-resource-manifest');

    cy.getByDataCy('resource-deploy-startup').first()
      .should('contain.text', 'Yes, forced');
  });

  it('displays a created resource, with disabled deploy', () => {
    cy.init();

    cy.selectTab(TAB_RESOURCES);
    cy.getByDataCy('resource-name').type('created-resource');
    cy.getByDataCy('resource-toggle-inlined').click();
    cy.getByDataCy('resource-manifest').type('a-resource-manifest');
    cy.getByDataCy('resource-auto-deploy-never').click();
    cy.getByDataCy('resource-create').click();

    cy.getByDataCy('resource-info').first()
      .should('contain.text', 'created-resource')
      .should('contain.text', 'a-resource-manifest');

    cy.getByDataCy('resource-deploy-startup').first()
      .should('contain.text', 'No, disabled');
  });

  it('displays an updated resource, with manifest', () => {
    cy.init();

    cy.selectTab(TAB_RESOURCES);
    cy.getByDataCy('resource-name').type('created-resource');
    cy.getByDataCy('resource-toggle-inlined').click();
    cy.getByDataCy('resource-manifest').type('a-resource-manifest');
    cy.getByDataCy('resource-create').click();

    cy.getByDataCy('resource-info').first()
      .should('contain.text', 'created-resource')
      .should('contain.text', 'a-resource-manifest');

    cy.getByDataCy('resource-edit').click();
    cy.getByDataCy('resource-manifest').type('{selectAll}{del}another-resource-manifest');
    cy.getByDataCy('resource-save').click();

    cy.getByDataCy('resource-info').first()
    .should('contain.text', 'created-resource')
    .should('contain.text', 'another-resource-manifest');
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

  it('displays a modified volume', () => {
    cy.init();

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-name').type('created-volume');
    cy.getByDataCy('volume-size').type('512Mi');
    cy.getByDataCy('volume-ephemeral').click();
    cy.getByDataCy('volume-create').click();

    cy.getByDataCy('volume-info').first()
      .should('contain.text', 'created-volume')
      .should('contain.text', '512Mi')
      .should('contain.text', 'Yes');

    cy.getByDataCy('volume-edit').click();
    cy.getByDataCy('volume-size').type('{selectAll}{del}1Gi');
    cy.getByDataCy('volume-ephemeral').click();
    cy.getByDataCy('volume-save').click();

    cy.getByDataCy('volume-info').first()
      .should('contain.text', 'created-volume')
      .should('contain.text', '1Gi')
      .should('contain.text', 'No');
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
    cy.getByDataCy('volume-mount-path-0').type("/mnt/vol1", {force: true});
    cy.getByDataCy('volume-mount-name-0').click().get('mat-option').contains('volume1').click();

    cy.getByDataCy('volume-mount-add').click();
    cy.getByDataCy('volume-mount-path-1').type("/mnt/vol2", {force: true});
    cy.getByDataCy('volume-mount-name-1').click().get('mat-option').contains('(New Volume)').click();
    cy.getByDataCy('volume-name').type('volume2');
    cy.getByDataCy('volume-create').click();
    
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
      .should('contain.text', '/mnt/vol1')
      .should('contain.text', 'volume2')
      .should('contain.text', '/mnt/vol2');

    cy.selectTab(TAB_VOLUMES);
    cy.getByDataCy('volume-info').eq(1)
      .should('contain.text', 'volume2');
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

    cy.getByDataCy('image-build-startup').first()
      .should('contain.text', 'No, the image is referenced by a command');
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

    cy.getByDataCy('resource-deploy-startup').first()
      .should('contain.text', 'No, the resource is referenced by a command');
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
