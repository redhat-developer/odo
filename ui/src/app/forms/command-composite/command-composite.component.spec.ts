import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CommandCompositeComponent } from './command-composite.component';

describe('CommandCompositeComponent', () => {
  let component: CommandCompositeComponent;
  let fixture: ComponentFixture<CommandCompositeComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CommandCompositeComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CommandCompositeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
