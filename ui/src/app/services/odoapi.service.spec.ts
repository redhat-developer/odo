import { TestBed } from '@angular/core/testing';

import { OdoapiService } from './odoapi.service';

describe('OdoapiService', () => {
  let service: OdoapiService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(OdoapiService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
