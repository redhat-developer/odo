import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CommandExecComponent } from './command-exec.component';

describe('CommandExecComponent', () => {
  let component: CommandExecComponent;
  let fixture: ComponentFixture<CommandExecComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CommandExecComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CommandExecComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
