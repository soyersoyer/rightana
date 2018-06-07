import { Directive } from '@angular/core';
import { FormGroupDirective } from '@angular/forms';

export function markAsTouchedDeep(control: any): void {
  control.markAsTouched();
  const controls = control.controls;
  if (!controls) {
    return;
  }
  if (controls instanceof Array) {
    controls.forEach(c => markAsTouchedDeep(c));
  }
  else {
    Object.keys(control.controls).forEach(key => markAsTouchedDeep(control.controls[key]));
  }
}

@Directive({ selector: '[mark-as-touched]' })
export class MarkAsToucedDirective {
  constructor(private fgDirective: FormGroupDirective) {
    this.fgDirective.ngSubmit.forEach(_ => markAsTouchedDeep(this.fgDirective.form));
  }
}
