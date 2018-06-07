import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { ToastyService } from './toasty.service';
import { ToastyComponent } from './toasty.component';
import { ToastComponent } from './toast.component';

export * from './toasty.service';

@NgModule({
  imports: [
    CommonModule
  ],
  declarations: [
    ToastyComponent,
    ToastComponent,
  ],
  providers: [
    ToastyService,
  ],
  exports: [
    ToastComponent,
    ToastyComponent,
  ],
})
export class ToastyModule { }
