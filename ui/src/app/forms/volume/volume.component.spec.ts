import { ComponentFixture, TestBed } from '@angular/core/testing';

import { VolumeComponent } from './volume.component';

describe('VolumeComponent', () => {
  let component: VolumeComponent;
  let fixture: ComponentFixture<VolumeComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ VolumeComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(VolumeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
