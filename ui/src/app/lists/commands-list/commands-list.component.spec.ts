import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CommandsListComponent } from './commands-list.component';

describe('CommandsListComponent', () => {
  let component: CommandsListComponent;
  let fixture: ComponentFixture<CommandsListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CommandsListComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CommandsListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
