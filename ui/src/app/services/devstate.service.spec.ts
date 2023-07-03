import { TestBed } from '@angular/core/testing';

import { DevstateService } from './devstate.service';

describe('DevstateService', () => {
  let service: DevstateService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(DevstateService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
