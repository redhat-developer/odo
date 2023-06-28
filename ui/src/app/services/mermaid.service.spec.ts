import { TestBed } from '@angular/core/testing';

import { MermaidService } from './mermaid.service';

describe('MermaidService', () => {
  let service: MermaidService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MermaidService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
