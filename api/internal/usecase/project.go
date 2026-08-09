package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pmo-agent/api/internal/domain"
)

// ProjectUsecase はプロジェクトの管理とコード発行を担う。
type ProjectUsecase struct {
	projects ProjectRepository
	programs ProgramRepository
}

func NewProjectUsecase(projects ProjectRepository, programs ProgramRepository) *ProjectUsecase {
	return &ProjectUsecase{projects: projects, programs: programs}
}

type CreateProjectInput struct {
	Name        string
	Description string
	PMID        *int
	ApproverID  *int
	Vendor      string
	Budget      *int64
	StartDate   *time.Time
	EndDate     *time.Time
	SourceType  string
	SourceValue string
	CreatedBy   int
}

type UpdateProjectInput struct {
	Name        string
	Description string
	PMID        *int
	ApproverID  *int
	Vendor      string
	Budget      *int64
	StartDate   *time.Time
	EndDate     *time.Time
	Status      domain.ProjectStatus
	SourceType  string
	SourceValue string
}

// Get はスコープ内のプロジェクトのみ返す。担当外は存在秘匿のため ErrNotFound を返す
// （Critical Business Rule: 担当外PJの予算・ベンダーを漏らさない）。
func (uc *ProjectUsecase) Get(ctx context.Context, id int, scope domain.ProjectScope) (*domain.Project, error) {
	if !scope.Allows(id) {
		return nil, domain.ErrNotFound
	}
	return uc.projects.FindByID(ctx, id)
}

// List はスコープ内のプロジェクトのみ返す。All=false かつ許可PJが空なら空配列。
func (uc *ProjectUsecase) List(ctx context.Context, scope domain.ProjectScope) ([]domain.Project, error) {
	if scope.All {
		return uc.projects.List(ctx)
	}
	if len(scope.ProjectIDs) == 0 {
		return []domain.Project{}, nil
	}
	return uc.projects.ListByIDs(ctx, scope.ProjectIDs)
}

// Create はプログラム配下にプロジェクトを起案する（status=planning, project_code=nil）。
func (uc *ProjectUsecase) Create(ctx context.Context, programID int, in CreateProjectInput) (*domain.Project, error) {
	if _, err := uc.programs.FindByID(ctx, programID); err != nil {
		return nil, err // プログラム未存在は ErrNotFound
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: プロジェクト名は必須です", domain.ErrValidation)
	}
	sourceType, sourceValue := normalizeSource(in.SourceType, in.SourceValue)
	p := &domain.Project{
		ProgramID:   programID,
		Name:        strings.TrimSpace(in.Name),
		Description: in.Description,
		PMID:        in.PMID,
		ApproverID:  in.ApproverID,
		Vendor:      in.Vendor,
		Budget:      in.Budget,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
		Status:      domain.StatusPlanning,
		SourceType:  sourceType,
		SourceValue: sourceValue,
		CreatedBy:   in.CreatedBy,
	}
	if err := uc.projects.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("usecase.Project.Create: %w", err)
	}
	return p, nil
}

func (uc *ProjectUsecase) Update(ctx context.Context, id int, in UpdateProjectInput) (*domain.Project, error) {
	p, err := uc.projects.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: プロジェクト名は必須です", domain.ErrValidation)
	}
	if !domain.IsValidStatus(in.Status) {
		return nil, fmt.Errorf("%w: ステータスが不正です", domain.ErrValidation)
	}
	if !domain.CanTransitionTo(p.Status, in.Status) {
		return nil, fmt.Errorf("%w: %s から %s へは遷移できません", domain.ErrConflict, p.Status, in.Status)
	}
	p.Name = strings.TrimSpace(in.Name)
	p.Description = in.Description
	p.PMID = in.PMID
	p.ApproverID = in.ApproverID
	p.Vendor = in.Vendor
	p.Budget = in.Budget
	p.StartDate = in.StartDate
	p.EndDate = in.EndDate
	p.Status = in.Status
	p.SourceType, p.SourceValue = normalizeSource(in.SourceType, in.SourceValue)
	if err := uc.projects.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("usecase.Project.Update: %w", err)
	}
	return uc.projects.FindByID(ctx, id)
}

func (uc *ProjectUsecase) Delete(ctx context.Context, id int) error {
	if _, err := uc.projects.FindByID(ctx, id); err != nil {
		return err
	}
	return uc.projects.Delete(ctx, id)
}

// IssueCode はプロジェクトコード（枝番付き）を発行し active へ遷移させる。
// 発行済み（project_code != nil）または planning 以外は拒否（発行後不変・一度きり）。
func (uc *ProjectUsecase) IssueCode(ctx context.Context, id int) (*domain.Project, error) {
	p, err := uc.projects.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.ProjectCode != nil {
		return nil, fmt.Errorf("%w: このプロジェクトは既にコードが発行されています", domain.ErrConflict)
	}
	if p.Status != domain.StatusPlanning {
		return nil, fmt.Errorf("%w: planning 状態のプロジェクトのみ発行できます", domain.ErrConflict)
	}
	program, err := uc.programs.FindByID(ctx, p.ProgramID)
	if err != nil {
		return nil, err
	}
	maxBranch, err := uc.projects.MaxBranchNo(ctx, p.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("usecase.Project.IssueCode max: %w", err)
	}
	branch := maxBranch + 1
	code := fmt.Sprintf("%s-%03d", program.Code, branch)
	if err := uc.projects.IssueCode(ctx, id, branch, code); err != nil {
		return nil, fmt.Errorf("usecase.Project.IssueCode persist: %w", err)
	}
	return uc.projects.FindByID(ctx, id)
}

// normalizeSource は進捗の取得元指定（種別と識別子）を組で正規化する。
// 種別が未指定なら識別子も捨てる。両者を独立に空文字判定すると
// source_type=NULL / source_value あり という行ができてしまい、収集対象の抽出が
// source_type で絞る以上、どのワークフローからも到達できない値だけが residue として残る。
// 逆に種別だけの指定は許す（先に取得元を決めて識別子を後から入れる運用があるため）。
func normalizeSource(sourceType, sourceValue string) (*string, *string) {
	t := emptyToNil(sourceType)
	if t == nil {
		return nil, nil
	}
	return t, emptyToNil(sourceValue)
}

// emptyToNil は空文字（前後空白のみを含む）を nil に正規化する。
// source_type は data_source_types.code への FK を持つため、未選択を空文字のまま
// 保存すると外部キー制約に違反する。
func emptyToNil(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
