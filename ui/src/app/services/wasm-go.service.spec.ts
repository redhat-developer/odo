import { TestBed } from '@angular/core/testing';

import { WasmGoService } from './wasm-go.service';

describe('WasmGoService', () => {
  let service: WasmGoService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(WasmGoService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
