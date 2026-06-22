import {useEffect, useMemo, useState} from 'react';
import {
    Alert,
    Badge,
    Button,
    Card,
    Col,
    Descriptions,
    Divider,
    Drawer,
    Form,
    Input,
    InputNumber,
    Layout,
    List,
    Menu,
    Progress,
    Row,
    Select,
    Slider,
    Space,
    Spin,
    Switch,
    Table,
    Tabs,
    Tag,
    Typography,
    message,
} from 'antd';
import {
    ApiOutlined,
    ExportOutlined,
    GithubOutlined,
    ReloadOutlined,
    SaveOutlined,
    SettingOutlined,
    ThunderboltOutlined,
} from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import 'antd/dist/reset.css';
import './App.css';
import {
    AnalyzeRepository,
    AskRepositoryQuestion,
    DiscoverProjects,
    GetResearchSession,
    ListResearchSessions,
    SaveSettings,
    Settings,
    TestGitHubConnectionDraft,
    TestLLMConnectionDraft,
} from '../wailsjs/go/main/App';
import {config, domain} from '../wailsjs/go/models';

const {Header, Content} = Layout;
const {Text, Title, Paragraph} = Typography;

const providerKeys = ['openai', 'anthropic', 'deepseek', 'custom'] as const;
type ProviderKey = typeof providerKeys[number];

const providerLabels: Record<ProviderKey, string> = {
    openai: 'OpenAI',
    anthropic: 'Claude / Anthropic',
    deepseek: 'DeepSeek',
    custom: '自定义中转',
};

const providerDefaults: Record<ProviderKey, { model: string; baseURL: string; enabled: boolean }> = {
    openai: {model: 'gpt-4.1-mini', baseURL: 'https://api.openai.com/v1', enabled: true},
    anthropic: {model: 'claude-3-5-sonnet-latest', baseURL: 'https://api.anthropic.com', enabled: true},
    deepseek: {model: 'deepseek-chat', baseURL: 'https://api.deepseek.com/v1', enabled: true},
    custom: {model: 'gpt-4.1-mini', baseURL: '', enabled: false},
};

const languageOptions = [
    {label: '自动', value: 'auto'},
    {label: '中文', value: 'zh'},
    {label: 'English', value: 'en'},
];

const uiLanguageOptions = [
    {label: '中文', value: 'zh'},
    {label: 'English', value: 'en'},
];

const repoLanguageOptions = [
    {label: 'Go', value: 'Go'},
    {label: 'Python', value: 'Python'},
    {label: 'TypeScript', value: 'TypeScript'},
    {label: 'JavaScript', value: 'JavaScript'},
    {label: 'Rust', value: 'Rust'},
    {label: 'Java', value: 'Java'},
];

const directionOptions = [
    {label: 'Agent / 智能体', value: 'agent'},
    {label: 'MCP / 工具协议', value: 'mcp'},
    {label: 'RAG / 检索增强', value: 'rag'},
    {label: 'Backend / 后端', value: 'backend'},
    {label: 'Framework / 框架', value: 'framework'},
];

const difficultyOptions = [
    {label: '入门', value: 'beginner'},
    {label: '进阶', value: 'intermediate'},
    {label: '高阶', value: 'advanced'},
];

type Lang = 'zh' | 'en';
type SettingsSection = 'github' | 'llm' | 'defaults' | 'storage';

const copy = {
    zh: {
        subtitle: '开源项目研究工作台：筛选、评分、单仓分析和贡献决策',
        settings: '设置',
        reload: '重新载入',
        save: '保存',
        saved: '配置已保存',
        githubReady: '已配置',
        githubMissing: '未配置',
        analysisTab: '单仓分析',
        scoreTab: '评分矩阵',
        reportTab: '研究报告',
        advancedFilters: '高级筛选',
        userQuery: '用户 Query',
        userQueryRequired: '请输入你的技术栈、方向和目标',
        repoLanguage: '仓库语言',
        inputLanguage: '输入语言',
        direction: '方向',
        customDirection: '自定义方向',
        customDirectionPlaceholder: '例如 observability / gateway / scheduler',
        difficulty: '难度',
        starRange: 'Star 范围',
        pushedAfter: '最近活跃',
        limit: '项目数量',
        start: '开始研究',
        queryHint: '中文和英文输入都会进入查询规划器。',
        starHint: 'Star 范围用于限制项目成熟度和贡献窗口。',
        llmHint: 'LLM 只用于总结和润色，评分仍保持确定性。',
        githubInfo: 'GitHub base URL 写入 config.yaml；token 只写入本地 secrets.json，也可用 GITHUB_TOKEN 环境变量覆盖。',
        githubBaseURL: 'GitHub API Base URL',
        githubBaseURLRequired: '请配置 GitHub API Base URL',
        githubToken: 'GitHub Token',
        keepToken: '留空表示保留当前已保存 token。',
        testGitHub: '检测 GitHub 配置',
        tokenReady: 'token 已配置',
        tokenMissing: 'token 未配置',
        llmProvider: 'LLM Provider',
        llmInfo: 'Provider 的模型、base URL、启用状态写入 YAML；API token 只写入 secrets.json，支持中转站和自定义网关。',
        currentProvider: '当前 Provider',
        currentModel: '当前模型',
        enabled: '启用',
        model: '模型',
        apiToken: 'API Token',
        test: '检测',
        defaults: '默认筛选',
        storage: '存储',
        uiLanguage: '界面语言',
        defaultInputLanguage: '默认输入语言',
        defaultRepoLanguage: '默认仓库语言',
        defaultDirection: '默认方向',
        defaultDifficulty: '默认难度',
        defaultPushedAfter: '默认最近活跃',
        defaultStarRange: '默认 Star 范围',
        storageWarning: 'config.yaml 不保存 token；secrets.json 使用 0600 权限保存本地密钥。生产级版本建议换成系统 Keychain。',
        queryPlanning: '查询规划',
        queryEmpty: '输入背景和目标后，会在这里显示可解释的 GitHub 查询。',
        repoRanking: '项目分级',
        repoEmpty: '暂无项目结果。建议先用较宽的 Star 范围和最近活跃时间启动搜索。',
        noDescription: '暂无描述',
        unknownLanguage: '未知语言',
        scoreTitle: '确定性评分拆解',
        scoreEmpty: '运行搜索后展示确定性评分拆解。',
        repository: '仓库',
        total: '总分',
        activity: '活跃',
        popularity: '热度',
        learning: '学习价值',
        contribution: '贡献友好',
        role: '角色相关',
        influence: '影响力',
        friendly: '友好度',
        openGitHub: '打开 GitHub',
        deepAnalysis: '深度分析',
        readmeSummary: 'README 摘要',
        issueCategory: 'Issue 分类',
        prRisk: 'PR 风险',
        docs: '文档',
        examples: '示例',
        tests: '测试',
        directoryTree: '目录树',
        contributionSignals: '贡献信号',
        dependencyFiles: '依赖文件',
        noDependency: '未识别到依赖文件。',
        emptyAnalysis: '从中间列表选择一个仓库，查看 README、目录树、依赖、Issue 和 PR 信号。',
        markdownReport: 'Markdown 报告',
        reportEmpty: '运行搜索后生成中文 Markdown 研究报告。',
        exportMd: '导出 .md',
    },
    en: {
        subtitle: 'GitHub open-source research workbench for filtering, scoring, repository analysis, and contribution decisions',
        settings: 'Settings',
        reload: 'Reload',
        save: 'Save',
        saved: 'Settings saved',
        githubReady: 'ready',
        githubMissing: 'missing',
        analysisTab: 'Repository',
        scoreTab: 'Scores',
        reportTab: 'Report',
        advancedFilters: 'Advanced Filters',
        userQuery: 'User Query',
        userQueryRequired: 'Enter your stack, direction, and goal',
        repoLanguage: 'Repository Language',
        inputLanguage: 'Input Language',
        direction: 'Direction',
        customDirection: 'Custom Direction',
        customDirectionPlaceholder: 'e.g. observability / gateway / scheduler',
        difficulty: 'Difficulty',
        starRange: 'Star Range',
        pushedAfter: 'Pushed After',
        limit: 'Result Count',
        start: 'Start Research',
        queryHint: 'Chinese and English inputs are both normalized by the query planner.',
        starHint: 'Star range controls project maturity and contribution window.',
        llmHint: 'LLM is used only for summaries and wording; deterministic scoring stays unchanged.',
        githubInfo: 'GitHub base URL is saved to config.yaml; token is saved only to local secrets.json or overridden by GITHUB_TOKEN.',
        githubBaseURL: 'GitHub API Base URL',
        githubBaseURLRequired: 'Configure GitHub API Base URL',
        githubToken: 'GitHub Token',
        keepToken: 'Leave blank to keep the saved token.',
        testGitHub: 'Test GitHub Config',
        tokenReady: 'token configured',
        tokenMissing: 'token missing',
        llmProvider: 'LLM Provider',
        llmInfo: 'Provider model, base URL, and enabled state are saved to YAML; API token is saved only to secrets.json.',
        currentProvider: 'Current Provider',
        currentModel: 'Current Model',
        enabled: 'Enabled',
        model: 'Model',
        apiToken: 'API Token',
        test: 'Test',
        defaults: 'Defaults',
        storage: 'Storage',
        uiLanguage: 'UI Language',
        defaultInputLanguage: 'Default Input Language',
        defaultRepoLanguage: 'Default Repository Language',
        defaultDirection: 'Default Direction',
        defaultDifficulty: 'Default Difficulty',
        defaultPushedAfter: 'Default Pushed After',
        defaultStarRange: 'Default Star Range',
        storageWarning: 'config.yaml never stores tokens; secrets.json stores local secrets with 0600 permissions. Production should use system Keychain.',
        queryPlanning: 'Query Planning',
        queryEmpty: 'Enter your background and goal to generate explainable GitHub queries.',
        repoRanking: 'Repository Ranking',
        repoEmpty: 'No results yet. Start with a broad star range and recent activity filter.',
        noDescription: 'No description',
        unknownLanguage: 'Unknown language',
        scoreTitle: 'Deterministic Score Breakdown',
        scoreEmpty: 'Run discovery to show deterministic scores.',
        repository: 'Repository',
        total: 'Total',
        activity: 'Activity',
        popularity: 'Popularity',
        learning: 'Learning',
        contribution: 'Contribution',
        role: 'Role',
        influence: 'Influence',
        friendly: 'Friendly',
        openGitHub: 'Open GitHub',
        deepAnalysis: 'Deep Analysis',
        readmeSummary: 'README Summary',
        issueCategory: 'Issue Categories',
        prRisk: 'PR Risk',
        docs: 'Docs',
        examples: 'Examples',
        tests: 'Tests',
        directoryTree: 'Directory Tree',
        contributionSignals: 'Contribution Signals',
        dependencyFiles: 'Dependency Files',
        noDependency: 'No dependency files detected.',
        emptyAnalysis: 'Select a repository from the middle list to inspect README, tree, dependencies, Issue, and PR signals.',
        markdownReport: 'Markdown Report',
        reportEmpty: 'Run discovery to generate a Markdown research report.',
        exportMd: 'Export .md',
    },
} as const;

function App() {
    const [searchForm] = Form.useForm<SearchFormValues>();
    const [settingsForm] = Form.useForm<SettingsFormValues>();
    const [result, setResult] = useState<domain.DiscoveryResult | null>(null);
    const [analysis, setAnalysis] = useState<domain.RepositoryAnalysis | null>(null);
    const [sessions, setSessions] = useState<domain.ResearchSession[]>([]);
    const [qaTurns, setQaTurns] = useState<domain.RepositoryQuestionTurn[]>([]);
    const [bundle, setBundle] = useState<config.SettingsBundle | null>(null);
    const [loading, setLoading] = useState(false);
    const [analysisLoading, setAnalysisLoading] = useState(false);
    const [qaLoading, setQaLoading] = useState(false);
    const [settingsLoading, setSettingsLoading] = useState(false);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [settingsSection, setSettingsSection] = useState<SettingsSection>('github');
    const [error, setError] = useState('');
    const uiLanguage = (bundle?.config.ui_language === 'en' ? 'en' : 'zh') as Lang;
    const text = copy[uiLanguage];

    useEffect(() => {
        void refreshSettings();
        void refreshResearchSessions();
    }, []);

    async function refreshSettings() {
        setSettingsLoading(true);
        try {
            const response = await Settings();
            const nextBundle = config.SettingsBundle.createFrom(response);
            setBundle(nextBundle);
            settingsForm.setFieldsValue(mapSettingsToForm(nextBundle));
            searchForm.setFieldsValue(mapConfigToSearch(nextBundle.config));
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setSettingsLoading(false);
        }
    }

    async function saveSettings() {
        setSettingsLoading(true);
        setError('');
        try {
            const values = await settingsForm.validateFields();
            const response = await SaveSettings(mapSettingsFormToUpdate(values));
            const nextBundle = config.SettingsBundle.createFrom(response);
            setBundle(nextBundle);
            settingsForm.setFieldsValue(mapSettingsToForm(nextBundle));
            searchForm.setFieldsValue(mapConfigToSearch(nextBundle.config));
            setSettingsOpen(false);
            message.success(text.saved);
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setSettingsLoading(false);
        }
    }

    async function testGitHub() {
        const values = settingsForm.getFieldsValue();
        const response = await TestGitHubConnectionDraft(config.GitHubConnectionRequest.createFrom({
            base_url: values.githubBaseURL || 'https://api.github.com',
            token: values.githubToken || '',
        }));
        const check = config.ConnectionCheck.createFrom(response);
        (check.ok ? message.success : message.warning)(check.message);
    }

    async function testProvider(provider: ProviderKey) {
        const values = settingsForm.getFieldsValue();
        const response = await TestLLMConnectionDraft(config.ProviderConnectionRequest.createFrom({
            provider,
            model: values[`${provider}Model`] || providerDefaults[provider].model,
            base_url: values[`${provider}BaseURL`] || providerDefaults[provider].baseURL,
            token: values[`${provider}Token`] || '',
            enabled: Boolean(values[`${provider}Enabled`]),
        }));
        const check = config.ConnectionCheck.createFrom(response);
        (check.ok ? message.success : message.warning)(`${providerLabels[provider]}: ${check.message}`);
    }

    async function runDiscovery(values: SearchFormValues) {
        setLoading(true);
        setError('');
        setAnalysis(null);
        try {
            const response = await DiscoverProjects(mapSearchFormToRequest(values));
            const nextResult = domain.DiscoveryResult.createFrom(response);
            setResult(nextResult);
            if (nextResult.repositories.length > 0) {
                await loadAnalysis(nextResult.repositories[0].repository.full_name);
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setLoading(false);
        }
    }

    async function loadAnalysis(fullName: string) {
        setAnalysisLoading(true);
        setQaTurns([]);
        try {
            const response = await AnalyzeRepository(fullName);
            setAnalysis(domain.RepositoryAnalysis.createFrom(response));
            void refreshResearchSessions();
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setAnalysisLoading(false);
        }
    }

    async function refreshResearchSessions() {
        try {
            const response = await ListResearchSessions(20);
            setSessions((response || []).map((item) => domain.ResearchSession.createFrom(item)));
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        }
    }

    async function openResearchSession(id: number) {
        setAnalysisLoading(true);
        setQaTurns([]);
        setError('');
        try {
            const response = await GetResearchSession(id);
            const session = domain.ResearchSession.createFrom(response);
            setAnalysis(session.analysis);
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setAnalysisLoading(false);
        }
    }

    async function askRepositoryQuestion(question: string) {
        if (!selectedRepo || !question.trim()) {
            return;
        }
        setQaLoading(true);
        setError('');
        try {
            const response = await AskRepositoryQuestion(domain.RepositoryQuestionRequest.createFrom({
                full_name: selectedRepo,
                question: question.trim(),
                history: qaTurns,
            }));
            const answer = domain.RepositoryQuestionResponse.createFrom(response);
            setQaTurns((turns) => [...turns, domain.RepositoryQuestionTurn.createFrom({
                question: answer.question,
                answer: answer.answer,
            })]);
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setQaLoading(false);
        }
    }

    const rows = useMemo(() => result?.repositories ?? [], [result]);
    const selectedRepo = analysis?.repository.full_name ?? rows[0]?.repository.full_name ?? '';

    return (
        <Layout className="app-shell">
            <Header className="app-header">
                <div className="header-title">
                    <Title level={3} className="app-title">trpc-GitHub-agent</Title>
                    <Text className="app-subtitle">{text.subtitle}</Text>
                </div>
                <Space wrap className="header-actions">
                    {bundle ? (
                        <>
                            <Tag color={bundle.status.active_provider_ready ? 'green' : 'gold'}>
                                {bundle.status.model_provider} / {bundle.status.model_name}
                            </Tag>
                            <Tag color={bundle.status.github_token_configured ? 'blue' : 'default'}>
                                GitHub {bundle.status.github_token_configured ? text.githubReady : text.githubMissing}
                            </Tag>
                        </>
                    ) : null}
                    <Button icon={<SettingOutlined/>} onClick={() => setSettingsOpen(true)}>{text.settings}</Button>
                </Space>
            </Header>

            <Content className="app-content">
                <div className="status-strip">
                    <Space wrap size={[8, 8]}>
                        <Tag color="blue">Deterministic scoring</Tag>
                        <Tag color={bundle?.status.active_provider_ready ? 'green' : 'gold'}>
                            {bundle?.status.active_provider_ready ? 'LLM summary ready' : 'LLM deterministic fallback'}
                        </Tag>
                        <Tag color="default">GitHub read-only tools</Tag>
                        <Tag color="purple">Research sessions</Tag>
                    </Space>
                </div>
                {error && <Alert type="error" message={error} showIcon className="top-alert"/>}
                {result?.warnings?.map((warning) => (
                    <Alert key={warning} type="warning" message={warning} showIcon className="top-alert"/>
                ))}

                <Row gutter={[16, 16]} className="workbench">
                    <Col xs={24} xl={6} xxl={5} className="workbench-column">
                        <SearchPanel form={searchForm} loading={loading} onFinish={runDiscovery} text={text}/>
                    </Col>

                    <Col xs={24} xl={9} xxl={9} className="workbench-column">
                        <Spin spinning={loading}>
                            <Space direction="vertical" size={12} className="full-width">
                                <QueryList result={result} text={text}/>
                                <RepositoryList rows={rows} onSelect={loadAnalysis} selected={selectedRepo} text={text}/>
                            </Space>
                        </Spin>
                    </Col>

                    <Col xs={24} xl={9} xxl={10} className="workbench-column">
                        <Spin spinning={loading || analysisLoading}>
                            <Tabs
                                className="detail-tabs"
                                items={[
                                    {
                                        key: 'analysis',
                                        label: text.analysisTab,
                                        children: analysis ? <RepositoryAnalysisView analysis={analysis} text={text}/> : <EmptyAnalysisHint text={text}/>,
                                    },
                                    {
                                        key: 'scores',
                                        label: text.scoreTab,
                                        children: <ScoreTable rows={rows} text={text}/>,
                                    },
                                    {
                                        key: 'report',
                                        label: text.reportTab,
                                        children: <ReportPreview result={result} text={text}/>,
                                    },
                                    {
                                        key: 'qa',
                                        label: '研究问答',
                                        children: <RepositoryQA selectedRepo={selectedRepo} turns={qaTurns} loading={qaLoading} onAsk={askRepositoryQuestion}/>,
                                    },
                                    {
                                        key: 'sessions',
                                        label: '研究会话',
                                        children: <ResearchSessions sessions={sessions} onReload={refreshResearchSessions} onOpen={openResearchSession}/>,
                                    },
                                ]}
                            />
                        </Spin>
                    </Col>
                </Row>
            </Content>

            <Drawer
                open={settingsOpen}
                onClose={() => setSettingsOpen(false)}
                title={text.settings}
                width={680}
                className="settings-drawer"
                extra={
                    <Space>
                        <Button onClick={refreshSettings} loading={settingsLoading} icon={<ReloadOutlined/>}>{text.reload}</Button>
                        <Button type="primary" onClick={saveSettings} loading={settingsLoading} icon={<SaveOutlined/>}>{text.save}</Button>
                    </Space>
                }
            >
                <Spin spinning={settingsLoading}>
                    <Form form={settingsForm} layout="vertical">
                        <div className="settings-layout">
                            <Menu
                                mode="inline"
                                selectedKeys={[settingsSection]}
                                onClick={({key}) => setSettingsSection(key as SettingsSection)}
                                className="settings-menu"
                                items={[
                                    {key: 'github', label: 'GitHub'},
                                    {key: 'llm', label: text.llmProvider},
                                    {key: 'defaults', label: text.defaults},
                                    {key: 'storage', label: text.storage},
                                ]}
                            />
                            <div className="settings-content">
                                {settingsSection === 'github' ? <GitHubSettings bundle={bundle} onTest={testGitHub} text={text}/> : null}
                                {settingsSection === 'llm' ? <ProviderSettings bundle={bundle} onTest={testProvider} text={text}/> : null}
                                {settingsSection === 'defaults' ? <DefaultSettings text={text}/> : null}
                                {settingsSection === 'storage' ? <StorageSettings bundle={bundle} text={text}/> : null}
                            </div>
                        </div>
                    </Form>
                </Spin>
            </Drawer>
        </Layout>
    );
}

function SearchPanel({form, loading, onFinish, text}: {
    form: ReturnType<typeof Form.useForm<SearchFormValues>>[0];
    loading: boolean;
    onFinish: (values: SearchFormValues) => void;
    text: typeof copy[Lang];
}) {
    return (
        <Card title={text.advancedFilters} className="panel compact-panel search-panel">
            <Form form={form} layout="vertical" initialValues={defaultSearchValues} onFinish={onFinish}>
                <Form.Item label={text.userQuery} name="userInput" rules={[{required: true, message: text.userQueryRequired}]}>
                    <Input.TextArea rows={5} placeholder="我是 Go 后端，想找 Agent/MCP 项目，适合秋招和开源贡献。"/>
                </Form.Item>
                <Row gutter={10}>
                    <Col span={12}>
                        <Form.Item label={text.repoLanguage} name="language">
                            <Select options={repoLanguageOptions} allowClear/>
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item label={text.inputLanguage} name="inputLanguage">
                            <Select options={languageOptions}/>
                        </Form.Item>
                    </Col>
                </Row>
                <Row gutter={10}>
                    <Col span={12}>
                        <Form.Item label={text.direction} name="direction">
                            <Select options={directionOptions} showSearch allowClear/>
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item label={text.difficulty} name="difficulty">
                            <Select options={difficultyOptions} allowClear/>
                        </Form.Item>
                    </Col>
                </Row>
                <Form.Item label={text.customDirection} name="directionCustom">
                    <Input placeholder={text.customDirectionPlaceholder}/>
                </Form.Item>
                <Form.Item label={text.starRange} name="starRange">
                    <Slider range min={0} max={100000} step={100} marks={{0: '0', 10000: '1万', 50000: '5万', 100000: '10万'}}/>
                </Form.Item>
                <Row gutter={10}>
                    <Col span={12}>
                        <Form.Item label={text.pushedAfter} name="pushedAfter">
                            <Input placeholder="2025-01-01"/>
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item label={text.limit} name="limit">
                            <InputNumber min={3} max={20} className="full-width"/>
                        </Form.Item>
                    </Col>
                </Row>
                <Button type="primary" htmlType="submit" icon={<ThunderboltOutlined/>} loading={loading} block>
                    {text.start}
                </Button>
            </Form>
            <Divider/>
            <Space direction="vertical" size={6} className="full-width subtle-list">
                <Text type="secondary">{text.queryHint}</Text>
                <Text type="secondary">{text.starHint}</Text>
                <Text type="secondary">{text.llmHint}</Text>
            </Space>
        </Card>
    );
}

function GitHubSettings({bundle, onTest, text}: { bundle: config.SettingsBundle | null; onTest: () => void; text: typeof copy[Lang] }) {
    return (
        <Space direction="vertical" size={14} className="full-width">
            <Alert type="info" showIcon message={text.githubInfo}/>
            <Form.Item label={text.githubBaseURL} name="githubBaseURL" rules={[{required: true, message: text.githubBaseURLRequired}]}>
                <Input placeholder="https://api.github.com"/>
            </Form.Item>
            <Form.Item label={text.githubToken} name="githubToken" tooltip={text.keepToken}>
                <Input.Password placeholder="ghp_... / github_pat_..."/>
            </Form.Item>
            <Space wrap>
                <Button icon={<ApiOutlined/>} onClick={onTest}>{text.testGitHub}</Button>
                {bundle ? (
                    <Badge
                        status={bundle.status.github_token_configured ? 'success' : 'warning'}
                        text={bundle.status.github_token_configured ? text.tokenReady : text.tokenMissing}
                    />
                ) : null}
            </Space>
        </Space>
    );
}

function ProviderSettings({bundle, onTest, text}: {
    bundle: config.SettingsBundle | null;
    onTest: (provider: ProviderKey) => void;
    text: typeof copy[Lang];
}) {
    return (
        <Space direction="vertical" size={14} className="full-width">
            <Alert type="info" showIcon message={text.llmInfo}/>
            <Row gutter={12}>
                <Col span={12}>
                    <Form.Item label={text.currentProvider} name="modelProvider">
                        <Select options={providerKeys.map((key) => ({label: providerLabels[key], value: key}))}/>
                    </Form.Item>
                </Col>
                <Col span={12}>
                    <Form.Item label={text.currentModel} name="modelName">
                        <Input placeholder="gpt-4.1-mini"/>
                    </Form.Item>
                </Col>
            </Row>
            {providerKeys.map((provider) => {
                const status = bundle?.status.providers?.find((item) => item.name === provider);
                return (
                    <Card
                        key={provider}
                        size="small"
                        className="provider-card"
                        title={
                            <Space>
                                <span>{providerLabels[provider]}</span>
                                <Badge status={status?.configured ? 'success' : 'default'} text={status?.configured ? text.tokenReady : text.tokenMissing}/>
                            </Space>
                        }
                        extra={<Button size="small" onClick={() => onTest(provider)}>{text.test}</Button>}
                    >
                        <Row gutter={10}>
                            <Col span={6}>
                                <Form.Item label={text.enabled} name={`${provider}Enabled`} valuePropName="checked">
                                    <Switch/>
                                </Form.Item>
                            </Col>
                            <Col span={9}>
                                <Form.Item label={text.model} name={`${provider}Model`}>
                                    <Input placeholder={providerDefaults[provider].model}/>
                                </Form.Item>
                            </Col>
                            <Col span={9}>
                                <Form.Item label={text.apiToken} name={`${provider}Token`} tooltip={text.keepToken}>
                                    <Input.Password placeholder="sk-..."/>
                                </Form.Item>
                            </Col>
                        </Row>
                        <Form.Item label="Base URL" name={`${provider}BaseURL`}>
                            <Input placeholder={providerDefaults[provider].baseURL || 'https://your-relay.example.com/v1'}/>
                        </Form.Item>
                    </Card>
                );
            })}
        </Space>
    );
}

function DefaultSettings({text}: { text: typeof copy[Lang] }) {
    return (
        <Space direction="vertical" size={14} className="full-width">
            <Row gutter={12}>
                <Col span={12}>
                    <Form.Item label={text.uiLanguage} name="uiLanguage">
                        <Select options={uiLanguageOptions}/>
                    </Form.Item>
                </Col>
                <Col span={12}>
                    <Form.Item label={text.defaultInputLanguage} name="inputLanguage">
                        <Select options={languageOptions}/>
                    </Form.Item>
                </Col>
            </Row>
            <Row gutter={12}>
                <Col span={12}>
                    <Form.Item label={text.defaultRepoLanguage} name="defaultLanguage">
                        <Select options={repoLanguageOptions}/>
                    </Form.Item>
                </Col>
                <Col span={12}>
                    <Form.Item label={text.defaultDirection} name="defaultDirection">
                        <Select options={directionOptions} showSearch/>
                    </Form.Item>
                </Col>
            </Row>
            <Row gutter={12}>
                <Col span={12}>
                    <Form.Item label={text.defaultDifficulty} name="defaultDifficulty">
                        <Select options={difficultyOptions}/>
                    </Form.Item>
                </Col>
                <Col span={12}>
                    <Form.Item label={text.defaultPushedAfter} name="defaultPushedAfter">
                        <Input placeholder="2025-01-01"/>
                    </Form.Item>
                </Col>
            </Row>
            <Form.Item label={text.defaultStarRange} name="defaultStarRange">
                <Slider range min={0} max={100000} step={100} marks={{0: '0', 10000: '1万', 50000: '5万', 100000: '10万'}}/>
            </Form.Item>
        </Space>
    );
}

function StorageSettings({bundle, text}: { bundle: config.SettingsBundle | null; text: typeof copy[Lang] }) {
    return (
        <Space direction="vertical" size={14} className="full-width">
            <Alert type="warning" showIcon message={text.storageWarning}/>
            <Descriptions column={1} size="small" bordered>
                <Descriptions.Item label="config.yaml">{bundle?.config_path || 'n/a'}</Descriptions.Item>
                <Descriptions.Item label="secrets.json">{bundle?.secrets_path || 'n/a'}</Descriptions.Item>
                <Descriptions.Item label="GitHub Base URL">{bundle?.status.github_base_url || 'n/a'}</Descriptions.Item>
            </Descriptions>
        </Space>
    );
}

type SearchFormValues = {
    userInput: string;
    limit: number;
    inputLanguage: string;
    language?: string;
    direction?: string;
    directionCustom?: string;
    starRange?: [number, number];
    pushedAfter?: string;
    difficulty?: string;
};

type SettingsFormValues = {
    modelProvider: ProviderKey;
    modelName: string;
    uiLanguage: string;
    githubBaseURL: string;
    defaultLanguage: string;
    inputLanguage: string;
    defaultDirection: string;
    defaultDifficulty: string;
    defaultStarRange: [number, number];
    defaultPushedAfter: string;
    githubToken: string;
} & Record<`${ProviderKey}Enabled`, boolean>
    & Record<`${ProviderKey}Model`, string>
    & Record<`${ProviderKey}BaseURL`, string>
    & Record<`${ProviderKey}Token`, string>;

function mapSearchFormToRequest(values: SearchFormValues): domain.SearchRequest {
    const starRange = values.starRange ?? [0, 0];
    const direction = values.directionCustom?.trim() || values.direction || '';
    return domain.SearchRequest.createFrom({
        user_input: values.userInput,
        limit: values.limit,
        input_language: values.inputLanguage,
        languages: values.language ? [values.language] : [],
        topics: direction ? [direction] : [],
        target_role: direction,
        difficulty: values.difficulty ?? '',
        direction,
        pushed_after: values.pushedAfter ?? '',
        min_stars: starRange[0] ?? 0,
        max_stars: starRange[1] ?? 0,
    });
}

function mapSettingsFormToUpdate(values: SettingsFormValues): config.SettingsUpdate {
    const providers = providerKeys.map((provider) => ({
        name: provider,
        model: values[`${provider}Model`] || providerDefaults[provider].model,
        base_url: values[`${provider}BaseURL`] || providerDefaults[provider].baseURL,
        enabled: Boolean(values[`${provider}Enabled`]),
    }));
    const providerTokens = providerKeys.reduce<Record<string, string>>((acc, provider) => {
        const token = values[`${provider}Token`];
        if (token) {
            acc[provider] = token;
        }
        return acc;
    }, {});
    const starRange = values.defaultStarRange ?? [100, 50000];
    return config.SettingsUpdate.createFrom({
        config: {
            model_provider: values.modelProvider,
            model_name: values.modelName,
            ui_language: values.uiLanguage,
            github_base_url: values.githubBaseURL,
            providers,
            default_language: values.defaultLanguage,
            input_language: values.inputLanguage,
            default_direction: values.defaultDirection,
            default_difficulty: values.defaultDifficulty,
            default_min_stars: starRange[0],
            default_max_stars: starRange[1],
            default_pushed_after: values.defaultPushedAfter,
            theme: 'dark',
        },
        github_token: values.githubToken,
        provider_tokens: providerTokens,
    });
}

function mapSettingsToForm(bundle: config.SettingsBundle): Partial<SettingsFormValues> {
    const providerValues = providerKeys.reduce<Record<string, string | boolean>>((acc, provider) => {
        const saved = bundle.config.providers?.find((item) => item.name === provider);
        acc[`${provider}Enabled`] = saved?.enabled ?? providerDefaults[provider].enabled;
        acc[`${provider}Model`] = saved?.model ?? providerDefaults[provider].model;
        acc[`${provider}BaseURL`] = saved?.base_url ?? providerDefaults[provider].baseURL;
        acc[`${provider}Token`] = '';
        return acc;
    }, {});
    return {
        modelProvider: normalizeProvider(bundle.config.model_provider),
        modelName: bundle.config.model_name,
        uiLanguage: bundle.config.ui_language || 'zh',
        githubBaseURL: bundle.config.github_base_url || 'https://api.github.com',
        defaultLanguage: bundle.config.default_language,
        inputLanguage: bundle.config.input_language,
        defaultDirection: bundle.config.default_direction,
        defaultDifficulty: bundle.config.default_difficulty,
        defaultStarRange: [bundle.config.default_min_stars, bundle.config.default_max_stars],
        defaultPushedAfter: bundle.config.default_pushed_after,
        githubToken: '',
        ...providerValues,
    };
}

function mapConfigToSearch(cfg: config.AppConfig): Partial<SearchFormValues> {
    return {
        language: cfg.default_language,
        inputLanguage: cfg.input_language,
        direction: cfg.default_direction,
        starRange: [cfg.default_min_stars, cfg.default_max_stars],
        pushedAfter: cfg.default_pushed_after,
        difficulty: cfg.default_difficulty,
    };
}

function QueryList({result, text}: { result: domain.DiscoveryResult | null; text: typeof copy[Lang] }) {
    if (!result) {
        return <Card className="empty-state"><Paragraph>{text.queryEmpty}</Paragraph></Card>;
    }
    return (
        <Card title={text.queryPlanning} className="panel compact-panel">
            <Space direction="vertical" size={12} className="full-width">
                <List
                    className="query-list"
                    dataSource={result.queries}
                    renderItem={(query) => (
                        <List.Item>
                            <Space direction="vertical" size={2} className="full-width">
                                <Text code className="query-code">{query.query}</Text>
                                <Text type="secondary">{query.reason}</Text>
                            </Space>
                        </List.Item>
                    )}
                />
                <TraceList steps={result.agent_trace || []}/>
            </Space>
        </Card>
    );
}

function RepositoryList({rows, onSelect, selected, text}: { rows: domain.ScoredRepository[]; onSelect: (fullName: string) => void; selected: string; text: typeof copy[Lang] }) {
    if (rows.length === 0) {
        return <Card className="empty-state"><Paragraph>{text.repoEmpty}</Paragraph></Card>;
    }
    return (
        <Card title={text.repoRanking} className="panel compact-panel repo-list-panel">
            <List
                dataSource={rows}
                renderItem={({repository, score}, index) => {
                    const active = selected === repository.full_name;
                    return (
                        <List.Item onClick={() => onSelect(repository.full_name)} className={active ? 'repo-item repo-item-active' : 'repo-item'}>
                            <Space direction="vertical" size={7} className="full-width">
                                <div className="repo-title-row">
                                    <Space size={8} className="repo-name">
                                        <Tag color="processing">#{index + 1}</Tag>
                                        <GithubOutlined/>
                                        <Text strong ellipsis={{tooltip: repository.full_name}}>{repository.full_name}</Text>
                                    </Space>
                                    <Tag color={levelColor(score.influence_level)}>{score.total_score}</Tag>
                                </div>
                                <Text type="secondary" ellipsis={{tooltip: repository.description || text.noDescription}}>{repository.description || text.noDescription}</Text>
                                <Space wrap size={[6, 4]} className="repo-meta-row">
                                    <Tag color="blue">{repository.language || text.unknownLanguage}</Tag>
                                    <Tag>{repository.stars?.toLocaleString()} stars</Tag>
                                    <Tag>{repository.forks?.toLocaleString()} forks</Tag>
                                    <Tag color="purple">{difficultyLabel(score.difficulty)}</Tag>
                                </Space>
                                <Progress percent={score.total_score} showInfo={false} size="small"/>
                            </Space>
                        </List.Item>
                    );
                }}
            />
        </Card>
    );
}

function ScoreTable({rows, text}: { rows: domain.ScoredRepository[]; text: typeof copy[Lang] }) {
    if (rows.length === 0) {
        return <Card className="empty-state"><Paragraph>{text.scoreEmpty}</Paragraph></Card>;
    }
    return (
        <Card title={text.scoreTitle} className="panel compact-panel">
            <Table
                rowKey={(row) => row.repository.full_name}
                dataSource={rows}
                pagination={false}
                size="small"
                scroll={{x: 980}}
                columns={[
                    {
                        title: text.repository,
                        dataIndex: ['repository', 'full_name'],
                        fixed: 'left',
                        render: (_value, row) => <a href={row.repository.html_url} target="_blank" rel="noreferrer">{row.repository.full_name}</a>,
                    },
                    {title: text.total, dataIndex: ['score', 'total_score'], sorter: (a, b) => a.score.total_score - b.score.total_score},
                    {title: text.activity, dataIndex: ['score', 'activity_score']},
                    {title: text.popularity, dataIndex: ['score', 'popularity_score']},
                    {title: text.learning, dataIndex: ['score', 'learning_value_score']},
                    {title: text.contribution, dataIndex: ['score', 'contribution_friendliness_score']},
                    {title: text.role, dataIndex: ['score', 'role_relevance_score']},
                    {title: text.influence, dataIndex: ['score', 'influence_level']},
                    {title: text.friendly, dataIndex: ['score', 'beginner_friendliness']},
                ]}
            />
        </Card>
    );
}

function RepositoryAnalysisView({analysis, text}: { analysis: domain.RepositoryAnalysis; text: typeof copy[Lang] }) {
    const llmInsight = analysis.llm_insight;
    const hasLLMInsight = Boolean(llmInsight?.ai_generated);
    return (
        <Space direction="vertical" size={12} className="full-width">
            <Card className="panel compact-panel analysis-hero">
                <Space direction="vertical" size={10} className="full-width">
                    <div className="analysis-heading">
                        <div>
                            <Title level={4}>{analysis.repository.full_name}</Title>
                            <Text type="secondary">{analysis.repository.description || text.noDescription}</Text>
                        </div>
                        <Button href={analysis.repository.html_url} target="_blank" icon={<ExportOutlined/>}>{text.openGitHub}</Button>
                    </div>
                    <Space wrap className="repo-metrics">
                        <Tag color="blue">{analysis.repository.language || text.unknownLanguage}</Tag>
                        <Tag>{analysis.repository.stars?.toLocaleString()} stars</Tag>
                        <Tag>{analysis.repository.open_issues_count?.toLocaleString()} issues</Tag>
                    </Space>
                </Space>
            </Card>
            <Card title={text.deepAnalysis} className="panel compact-panel">
                <Descriptions column={1} size="small">
                    <Descriptions.Item label="项目定位">{analysis.positioning || text.noDescription}</Descriptions.Item>
                    <Descriptions.Item label="架构理解">{analysis.architecture || analysis.directory_summary}</Descriptions.Item>
                    <Descriptions.Item label={text.readmeSummary}>{analysis.profile.readme_summary || text.noDescription}</Descriptions.Item>
                    <Descriptions.Item label={text.issueCategory}>{analysis.issue_summary}</Descriptions.Item>
                    <Descriptions.Item label={text.prRisk}>{analysis.pr_summary}</Descriptions.Item>
                    <Descriptions.Item label={text.docs}>{analysis.docs_summary}</Descriptions.Item>
                    <Descriptions.Item label={text.examples}>{analysis.examples_summary}</Descriptions.Item>
                    <Descriptions.Item label={text.tests}>{analysis.tests_summary}</Descriptions.Item>
                    <Descriptions.Item label={text.directoryTree}>{analysis.directory_summary}</Descriptions.Item>
                </Descriptions>
            </Card>
            <Card title="结构化研究报告" className="panel compact-panel">
                <Row gutter={[12, 12]}>
                    <Col xs={24} md={12}>
                        <Space direction="vertical" size={8} className="full-width">
                            <Text strong>适合学习的模块</Text>
                            <Space wrap>
                                {(analysis.learning_modules || []).map((module) => (
                                    <Tag key={module} color="blue">{module}</Tag>
                                ))}
                            </Space>
                        </Space>
                    </Col>
                    <Col xs={24} md={12}>
                        <Space direction="vertical" size={8} className="full-width">
                            <Text strong>适合贡献的 Issue 类型</Text>
                            <Space wrap>
                                {(analysis.contribution_types || []).map((type) => (
                                    <Tag key={type} color={type === 'good-first' ? 'green' : 'default'}>{type}</Tag>
                                ))}
                            </Space>
                        </Space>
                    </Col>
                </Row>
            </Card>
            <Card title="AI 研究洞察" className="panel compact-panel">
                {hasLLMInsight ? (
                    <Space direction="vertical" size={10} className="full-width">
                        <Space wrap>
                            <Tag color="gold">AI generated</Tag>
                            <Tag>{llmInsight.provider} / {llmInsight.model}</Tag>
                            <Tag color="blue">不参与确定性评分</Tag>
                        </Space>
                        <Descriptions column={1} size="small">
                            <Descriptions.Item label="README 摘要">{llmInsight.readme_summary || analysis.profile.readme_summary || text.noDescription}</Descriptions.Item>
                            <Descriptions.Item label="Issue 解释">{llmInsight.issue_explanation || analysis.issue_summary}</Descriptions.Item>
                            <Descriptions.Item label="PR 风险总结">{llmInsight.pr_risk_summary || analysis.pr_summary}</Descriptions.Item>
                            <Descriptions.Item label="贡献路线">{llmInsight.contribution_plan || analysis.contribution_plan}</Descriptions.Item>
                            <Descriptions.Item label="推荐理由">{llmInsight.recommendation || text.noDescription}</Descriptions.Item>
                        </Descriptions>
                    </Space>
                ) : (
                    <Alert
                        type="info"
                        showIcon
                        message={llmInsight?.generation_warning || 'LLM provider 未启用，当前仅展示确定性工具链分析。'}
                        description="配置并检测 LLM provider 后，AnalyzeRepository 会补充 README 摘要、Issue 解释、PR 风险和贡献路线润色。"
                    />
                )}
            </Card>
            <Card title="Agent 执行轨迹" className="panel compact-panel agent-trace-card">
                <List
                    size="small"
                    dataSource={analysis.agent_trace || []}
                    renderItem={(step) => (
                        <List.Item className="trace-step">
                            <Space direction="vertical" size={2} className="full-width">
                                <Space wrap>
                                    <Tag color={step.phase === 'Plan' ? 'blue' : step.phase === 'Findings' ? 'green' : 'purple'}>{step.phase}</Tag>
                                    <Text code>{step.tool}</Text>
                                </Space>
                                <Text type="secondary">{step.summary}</Text>
                            </Space>
                        </List.Item>
                    )}
                />
            </Card>
            <Card title="贡献路线与简历价值" className="panel compact-panel">
                <Space direction="vertical" size={10} className="full-width">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>{analysis.contribution_plan || ''}</ReactMarkdown>
                    <Alert
                        type={hasLLMInsight ? 'success' : 'info'}
                        showIcon
                        message={hasLLMInsight ? 'AI generated 已启用：上方 AI 研究洞察为 LLM 补充；本区保留确定性贡献路线作对照。' : 'AI generated 未启用：当前为确定性工具链生成。'}
                    />
                    <Text>{analysis.resume_value}</Text>
                </Space>
            </Card>
            <Row gutter={[12, 12]}>
                <Col xs={24} md={12}>
                    <Card title={text.contributionSignals} className="panel compact-panel">
                        <Space wrap>
                            <Tag color={analysis.profile.has_docs ? 'green' : 'default'}>docs</Tag>
                            <Tag color={analysis.profile.has_examples ? 'green' : 'default'}>examples</Tag>
                            <Tag color={analysis.profile.has_tests ? 'green' : 'default'}>tests</Tag>
                            <Tag color={analysis.profile.has_contributing ? 'green' : 'default'}>contributing</Tag>
                        </Space>
                    </Card>
                </Col>
                <Col xs={24} md={12}>
                    <Card title={text.dependencyFiles} className="panel compact-panel">
                        {analysis.dependency_files.length > 0 ? (
                            <Space wrap>
                                {analysis.dependency_files.map((file) => <Tag key={file}>{file}</Tag>)}
                            </Space>
                        ) : (
                            <Text type="secondary">{text.noDependency}</Text>
                        )}
                    </Card>
                </Col>
            </Row>
        </Space>
    );
}

function EmptyAnalysisHint({text}: { text: typeof copy[Lang] }) {
    return <Card className="empty-state"><Paragraph>{text.emptyAnalysis}</Paragraph></Card>;
}

function ReportPreview({result, text}: { result: domain.DiscoveryResult | null; text: typeof copy[Lang] }) {
    const markdown = result?.markdown_report ?? '';
    if (!markdown) {
        return <Card className="empty-state"><Paragraph>{text.reportEmpty}</Paragraph></Card>;
    }
    return (
        <Card title={text.markdownReport} className="panel compact-panel report-panel" extra={<Button icon={<ExportOutlined/>} onClick={() => downloadMarkdown(markdown)}>{text.exportMd}</Button>}>
            <TraceList steps={(result?.agent_trace || []).filter((step) => step.phase === 'Report' || step.tool === 'generate_project_report')}/>
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{markdown}</ReactMarkdown>
        </Card>
    );
}

function TraceList({steps}: { steps: domain.AgentTraceStep[] }) {
    if (!steps || steps.length === 0) {
        return null;
    }
    return (
        <List
            size="small"
            className="inline-trace-list"
            dataSource={steps}
            renderItem={(step) => (
                <List.Item className="inline-trace-step">
                    <Space direction="vertical" size={2} className="full-width">
                        <Space wrap>
                            <Tag color={step.phase === 'Plan' ? 'blue' : step.phase === 'Findings' ? 'green' : step.phase === 'Report' ? 'gold' : 'purple'}>{step.phase}</Tag>
                            <Text code>{step.tool}</Text>
                        </Space>
                        <Text type="secondary">{step.summary}</Text>
                    </Space>
                </List.Item>
            )}
        />
    );
}

function RepositoryQA({selectedRepo, turns, loading, onAsk}: {
    selectedRepo: string;
    turns: domain.RepositoryQuestionTurn[];
    loading: boolean;
    onAsk: (question: string) => void;
}) {
    const [question, setQuestion] = useState('这个项目适合我吗？我应该先看哪个目录？');
    if (!selectedRepo) {
        return <Card className="empty-state"><Paragraph>先选择一个仓库，再向研究 Agent 提问。</Paragraph></Card>;
    }
    return (
        <Card title="研究 Agent 问答" className="panel compact-panel">
            <Space direction="vertical" size={12} className="full-width">
                <Alert type="info" showIcon message="AI generated：此区域调用当前 LLM provider，只用于解释、总结和建议；不参与确定性排序。"/>
                {turns.length > 0 ? (
                    <List
                        className="qa-thread"
                        dataSource={turns}
                        renderItem={(turn, index) => (
                            <List.Item className="qa-turn">
                                <Space direction="vertical" size={8} className="full-width">
                                    <Space wrap>
                                        <Tag color="blue">Q{index + 1}</Tag>
                                        <Text strong>{turn.question}</Text>
                                    </Space>
                                    <div className="qa-answer">
                                        <Tag color="gold">AI generated</Tag>
                                        <ReactMarkdown remarkPlugins={[remarkGfm]}>{turn.answer}</ReactMarkdown>
                                    </div>
                                </Space>
                            </List.Item>
                        )}
                    />
                ) : null}
                <Input.TextArea
                    rows={4}
                    value={question}
                    onChange={(event) => setQuestion(event.target.value)}
                    placeholder="例如：我应该先看哪个目录？哪些 issue 适合新手？怎么写第一版 PR？"
                />
                <Button
                    type="primary"
                    loading={loading}
                    onClick={() => {
                        const nextQuestion = question.trim();
                        if (!nextQuestion) {
                            return;
                        }
                        onAsk(nextQuestion);
                        setQuestion('');
                    }}
                    block
                >
                    向 Agent 提问
                </Button>
            </Space>
        </Card>
    );
}

function ResearchSessions({sessions, onReload, onOpen}: {
    sessions: domain.ResearchSession[];
    onReload: () => void;
    onOpen: (id: number) => void;
}) {
    if (sessions.length === 0) {
        return (
            <Card className="empty-state">
                <Space direction="vertical" size={12} className="full-width">
                    <Paragraph>还没有保存的研究会话。完成一次单仓分析后，会自动保存可回看的 Agent 报告。</Paragraph>
                    <Button icon={<ReloadOutlined/>} onClick={onReload}>刷新</Button>
                </Space>
            </Card>
        );
    }
    return (
        <Card
            title="研究会话"
            className="panel compact-panel"
            extra={<Button size="small" icon={<ReloadOutlined/>} onClick={onReload}>刷新</Button>}
        >
            <List
                size="small"
                dataSource={sessions}
                renderItem={(session) => (
                    <List.Item
                        className="session-item"
                        actions={[
                            <Button key="open" size="small" onClick={() => onOpen(session.id)}>回看</Button>,
                        ]}
                    >
                        <Space direction="vertical" size={4} className="full-width">
                            <Space wrap>
                                <Text strong>{session.repository}</Text>
                                <Tag color={session.ai_generated ? 'gold' : 'default'}>
                                    {session.ai_generated ? 'AI generated' : 'deterministic'}
                                </Tag>
                                {session.provider ? <Tag>{session.provider} / {session.model}</Tag> : null}
                            </Space>
                            <Space wrap size={[6, 4]}>
                                <Text type="secondary">{formatDateTime(session.created_at)}</Text>
                                <Text type="secondary">{session.trace_step_count} trace steps</Text>
                            </Space>
                        </Space>
                    </List.Item>
                )}
            />
        </Card>
    );
}

function downloadMarkdown(markdown: string) {
    const blob = new Blob([markdown], {type: 'text/markdown;charset=utf-8'});
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'github-research-report.md';
    a.click();
    URL.revokeObjectURL(url);
}

function formatDateTime(value: unknown) {
    if (!value) {
        return 'unknown time';
    }
    const date = new Date(String(value));
    if (Number.isNaN(date.getTime())) {
        return String(value);
    }
    return date.toLocaleString();
}

function normalizeProvider(provider: string): ProviderKey {
    if (provider === 'claude') {
        return 'anthropic';
    }
    return providerKeys.includes(provider as ProviderKey) ? provider as ProviderKey : 'openai';
}

function difficultyLabel(value: string) {
    const option = difficultyOptions.find((item) => item.value === value);
    return option?.label ?? value;
}

function levelColor(level: string) {
    switch (level) {
        case 'S':
            return 'red';
        case 'A':
            return 'volcano';
        case 'B':
            return 'blue';
        default:
            return 'default';
    }
}

const defaultSearchValues: SearchFormValues = {
    userInput: '我是 Go 后端，想找 Agent/MCP 项目，适合秋招和开源贡献。',
    limit: 5,
    inputLanguage: 'auto',
    language: 'Go',
    direction: 'agent',
    starRange: [100, 50000],
    pushedAfter: '2025-01-01',
    difficulty: 'intermediate',
};

export default App;
