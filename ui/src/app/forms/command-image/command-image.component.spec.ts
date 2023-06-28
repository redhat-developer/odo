import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CommandImageComponent } from './command-image.component';

describe('CommandImageComponent', () => {
  let component: CommandImageComponent;
  let fixture: ComponentFixture<CommandImageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CommandImageComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CommandImageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
