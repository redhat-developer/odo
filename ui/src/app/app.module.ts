import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ReactiveFormsModule } from '@angular/forms';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from "@angular/common/http";

import { DragDropModule } from '@angular/cdk/drag-drop';

import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatCardModule } from '@angular/material/card';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipsModule } from '@angular/material/chips';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { MatSelectModule } from '@angular/material/select';
import { MatTabsModule } from '@angular/material/tabs';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';

import { AppComponent } from './app.component';
import { MetadataComponent } from './forms/metadata/metadata.component';
import { MultiTextComponent } from './controls/multi-text/multi-text.component';
import { ContainersComponent } from './tabs/containers/containers.component';
import { ContainerComponent } from './forms/container/container.component';
import { CommandsComponent } from './tabs/commands/commands.component';
import { CommandExecComponent } from './forms/command-exec/command-exec.component';
import { CommandApplyComponent } from './forms/command-apply/command-apply.component';
import { CommandCompositeComponent } from './forms/command-composite/command-composite.component';
import { SelectContainerComponent } from './controls/select-container/select-container.component';
import { ResourcesComponent } from './tabs/resources/resources.component';
import { ResourceComponent } from './forms/resource/resource.component';
import { ImagesComponent } from './tabs/images/images.component';
import { ImageComponent } from './forms/image/image.component';
import { CommandImageComponent } from './forms/command-image/command-image.component';
import { CommandsListComponent } from './lists/commands-list/commands-list.component';
import { MultiCommandComponent } from './controls/multi-command/multi-command.component';
import { EventsComponent } from './tabs/events/events.component';
import { ChipsEventsComponent } from './controls/chips-events/chips-events.component';

@NgModule({
  declarations: [
    AppComponent,
    MetadataComponent,
    MultiTextComponent,
    ContainersComponent,
    ContainerComponent,
    CommandsComponent,
    CommandExecComponent,
    CommandApplyComponent,
    CommandCompositeComponent,
    SelectContainerComponent,
    ResourcesComponent,
    ResourceComponent,
    ImagesComponent,
    ImageComponent,
    CommandImageComponent,
    CommandsListComponent,
    MultiCommandComponent,
    EventsComponent,
    ChipsEventsComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    ReactiveFormsModule,
    FormsModule,
    HttpClientModule,
    
    DragDropModule,
    
    MatAutocompleteModule,
    MatButtonModule,
    MatButtonToggleModule,
    MatCardModule,
    MatCheckboxModule,
    MatChipsModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatMenuModule,
    MatSelectModule,
    MatTabsModule,
    MatToolbarModule,
    MatTooltipModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
