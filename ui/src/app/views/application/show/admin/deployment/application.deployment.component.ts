import {Component, Input, ViewChild} from '@angular/core';
import {TranslateService} from '@ngx-translate/core';
import {finalize, first} from 'rxjs/operators';
import {Application} from '../../../../../model/application.model';
import {ProjectPlatform} from '../../../../../model/platform.model';
import {Project} from '../../../../../model/project.model';
import {ApplicationStore} from '../../../../../service/application/application.store';
import {WarningModalComponent} from '../../../../../shared/modal/warning/warning.component';
import {ToastService} from '../../../../../shared/toast/ToastService';

@Component({
    selector: 'app-application-deployment',
    templateUrl: './application.deployment.html',
    styleUrls: ['./application.deployment.scss']
})
export class ApplicationDeploymentComponent {

    filteredPlatforms: Array<ProjectPlatform>;
    selectedPlatform: ProjectPlatform;

    public loadingBtn = false;

    _project: Project;

    @ViewChild('removeWarning') removeWarningModal: WarningModalComponent;
    @ViewChild('linkWarning') linkWarningModal: WarningModalComponent;

    @Input() application: Application;
    @Input('project')
    set project(project: Project) {
        this._project = project;
        if (project.platforms) {
            this.filteredPlatforms = project.platforms.filter(p => p.model.deployment);
        }
    }
    get project(): Project {
        return this._project;
    }

    getPlatformNames(): Array<string> {
        if (this.application.deployment_strategies) {
            return Object.keys(this.application.deployment_strategies);
        }
        return null;
    }

    constructor(private _appStore: ApplicationStore, private _toast: ToastService,
                public _translate: TranslateService) {
    }

    clickDeletePlatform(pfName: string) {
        this.loadingBtn = true;
        this._appStore.deleteDeploymentStrategy(
            this._project.key,
            this.application.name,
            pfName)
        .pipe(
            first(),
            finalize(() => this.loadingBtn = false)
        ).subscribe(
            app => {
                this.application = app;
                this._toast.success('', this._translate.instant('application_platform_deleted'));
            }
        );
    }

    updatePlatform(pfName: string) {
        this.loadingBtn = true;
        this._appStore.saveDeploymentStrategy(
            this._project.key,
            this.application.name,
            pfName,
            this.application.deployment_strategies[pfName])
        .pipe(
            first(),
            finalize(() => this.loadingBtn = false)
        ).subscribe(
            app => {
                this.application = app;
                this._toast.success('', this._translate.instant('application_platform_updated'));
            }
        );
    }

    addPlatform() {
        this.loadingBtn = true;
        if (this.selectedPlatform.model) {
            this._appStore.saveDeploymentStrategy(
                this._project.key,
                this.application.name,
                this.selectedPlatform.name,
                this.selectedPlatform.model.deployment_default_config)
            .pipe(
                first(),
                finalize(() => this.loadingBtn = false))
            .subscribe(
                app => {
                    this.application = app;
                    this.selectedPlatform = null;
                    this._toast.success('', this._translate.instant('application_platform_added'));
                }
            );
        }
    }
}
