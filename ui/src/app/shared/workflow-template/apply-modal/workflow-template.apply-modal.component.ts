import { Component, Input, ViewChild } from '@angular/core';
import { ModalTemplate, TemplateModalConfig } from 'ng2-semantic-ui';
import { ActiveModal, SuiModalService } from 'ng2-semantic-ui/dist';
import { forkJoin } from 'rxjs';
import { Project } from '../../../model/project.model';
import { WorkflowTemplate, WorkflowTemplateInstance } from '../../../model/workflow-template.model';
import { Workflow } from '../../../model/workflow.model';
import { WorkflowTemplateService } from '../../../service/services.module';
import { calculateWorkflowTemplateDiff } from '../../diff/diff';
import { Item } from '../../diff/list/diff.list.component';

@Component({
    selector: 'app-workflow-template-apply-modal',
    templateUrl: './workflow-template.apply-modal.html',
    styleUrls: ['./workflow-template.apply-modal.scss']
})
export class WorkflowTemplateApplyModalComponent {
    @ViewChild('workflowTemplateApplyModal') workflowTemplateApplyModal: ModalTemplate<boolean, boolean, void>;
    modal: ActiveModal<boolean, boolean, void>;
    open: boolean;

    @Input() project: Project;
    @Input() workflow: Workflow;
    workflowTemplate: WorkflowTemplate;
    workflowTemplateInstance: WorkflowTemplateInstance;
    diffVisible: boolean;
    diffItems: Array<Item>;
    workflowTemplateAuditMessages: Array<string>;

    constructor(
        private _modalService: SuiModalService,
        private _templateService: WorkflowTemplateService
    ) { }

    show() {
        if (this.open) {
            return;
        }

        this.open = true;

        const config = new TemplateModalConfig<boolean, boolean, void>(this.workflowTemplateApplyModal);
        config.mustScroll = true;
        this.modal = this._modalService.open(config);
        this.modal.onApprove(() => { this.open = false; });
        this.modal.onDeny(() => { this.open = false; });

        this.load();
    }

    load() {
        let s = this.workflow.from_template.split('/');

        forkJoin(this._templateService.get(s[0], s.splice(1, s.length - 1).join('/')),
            this._templateService.getInstance(this.project.key, this.workflow.name)).subscribe(res => {
                this.workflowTemplate = res[0];
                this.workflowTemplateInstance = res[1];

                // load audits since instance version if not latest
                if (this.workflowTemplateInstance.workflow_template_version !== this.workflowTemplate.version) {
                    this._templateService.getAudits(this.workflowTemplate.group.name, this.workflowTemplate.slug,
                        this.workflowTemplateInstance.workflow_template_version).subscribe(as => {
                            this.workflowTemplateAuditMessages = as.filter(a => !!a.change_message).map(a => a.change_message);
                            let before = as[0].data_after ? <WorkflowTemplate>JSON.parse(as[0].data_after) : null;
                            this.diffItems = calculateWorkflowTemplateDiff(before, this.workflowTemplate);
                        });
                } else {
                    this.workflowTemplateAuditMessages = [];
                    this.diffItems = [];
                }
            });
    }

    close() {
        this.diffVisible = false;
        this.modal.approve(true);
    }

    toggleDiff() {
        this.diffVisible = !this.diffVisible;
    }
}
