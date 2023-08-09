import { ComponentFixture, TestBed } from '@angular/core/testing';

import { VolumeMountsComponent } from './volume-mounts.component';

describe('VolumeMountsComponent', () => {
  let component: VolumeMountsComponent;
  let fixture: ComponentFixture<VolumeMountsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ VolumeMountsComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(VolumeMountsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
