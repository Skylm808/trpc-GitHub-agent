import {useMemo, useState} from 'react';
import {
    Alert,
    Button,
    Card,
    Col,
    Descriptions,
    Divider,
    Form,
    Input,
    InputNumber,
    Layout,
    List,
    Progress,
    Row,
    Space,
    Spin,
    Table,
    Tabs,
    Tag,
    Typography,
} from 'antd';
import {ExportOutlined, GithubOutlined, SearchOutlined} from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import 'antd/dist/reset.css';
import './App.css';
import {DiscoverProjects} from '../wailsjs/go/main/App';
import {domain} from '../wailsjs/go/models';

const {Header, Content} = Layout;
const {Text, Title, Paragraph} = Typography;
const defaultPrompt = '我是 Go 后端，帮我找 Go Agent 项目，适合秋招和开源贡献。';

function App() {
    const [form] = Form.useForm();
    const [result, setResult] = useState<domain.DiscoveryResult | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    async function runDiscovery(values: { prompt: string; limit: number }) {
        setLoading(true);
        setError('');
        try {
            const response = await DiscoverProjects(values.prompt, values.limit);
            setResult(domain.DiscoveryResult.createFrom(response));
        } catch (err) {
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setLoading(false);
        }
    }

    const rows = useMemo(() => result?.repositories ?? [], [result]);

    return (
        <Layout className="app-shell">
            <Header className="app-header">
                <div>
                    <Title level={3} className="app-title">trpc-GitHub-agent</Title>
                    <Text className="app-subtitle">GitHub open source research workbench for learning, recruiting, and contribution decisions</Text>
                </div>
                <Tag color={result?.used_live_github ? 'green' : 'gold'}>
                    {result?.used_live_github ? 'Live GitHub API' : 'Demo / cache mode'}
                </Tag>
            </Header>

            <Content className="app-content">
                <Row gutter={[20, 20]}>
                    <Col xs={24} lg={8}>
                        <Card title="Research Goal" className="panel">
                            <Form
                                form={form}
                                layout="vertical"
                                initialValues={{prompt: defaultPrompt, limit: 5}}
                                onFinish={runDiscovery}
                            >
                                <Form.Item
                                    label="目标描述"
                                    name="prompt"
                                    rules={[{required: true, message: '请输入你的技术栈、方向和目标'}]}
                                >
                                    <Input.TextArea rows={7} placeholder={defaultPrompt}/>
                                </Form.Item>
                                <Form.Item label="项目数量" name="limit">
                                    <InputNumber min={3} max={10} className="full-width"/>
                                </Form.Item>
                                <Button type="primary" htmlType="submit" icon={<SearchOutlined/>} loading={loading} block>
                                    Discover Projects
                                </Button>
                            </Form>
                        </Card>

                        {result && (
                            <Card title="Detected Intent" className="panel">
                                <Descriptions column={1} size="small">
                                    <Descriptions.Item label="Languages">{result.intent.languages?.join(', ')}</Descriptions.Item>
                                    <Descriptions.Item label="Topics">{result.intent.topics?.join(', ')}</Descriptions.Item>
                                    <Descriptions.Item label="Role">{result.intent.target_role}</Descriptions.Item>
                                    <Descriptions.Item label="Goals">{result.intent.goals?.join(', ')}</Descriptions.Item>
                                    <Descriptions.Item label="Difficulty">{result.intent.difficulty}</Descriptions.Item>
                                </Descriptions>
                            </Card>
                        )}
                    </Col>

                    <Col xs={24} lg={16}>
                        {error && <Alert type="error" message={error} className="panel"/>}
                        {result?.warnings?.map((warning) => (
                            <Alert key={warning} type="warning" message={warning} showIcon className="panel"/>
                        ))}

                        <Spin spinning={loading}>
                            <Tabs
                                items={[
                                    {
                                        key: 'projects',
                                        label: 'Projects',
                                        children: (
                                            <Space direction="vertical" size={16} className="full-width">
                                                <QueryList result={result}/>
                                                <ProjectCards rows={rows}/>
                                                <ScoreTable rows={rows}/>
                                            </Space>
                                        ),
                                    },
                                    {
                                        key: 'report',
                                        label: 'Markdown Report',
                                        children: <ReportPreview markdown={result?.markdown_report ?? ''}/>,
                                    },
                                ]}
                            />
                        </Spin>
                    </Col>
                </Row>
            </Content>
        </Layout>
    );
}

function QueryList({result}: { result: domain.DiscoveryResult | null }) {
    if (!result) {
        return (
            <Card className="empty-state">
                <Paragraph>Enter your background and target, then run discovery to generate explainable GitHub queries.</Paragraph>
            </Card>
        );
    }
    return (
        <Card title="Generated GitHub Queries" className="panel">
            <List
                dataSource={result.queries}
                renderItem={(query) => (
                    <List.Item>
                        <Space direction="vertical" size={4}>
                            <Text code>{query.query}</Text>
                            <Text type="secondary">{query.reason}</Text>
                        </Space>
                    </List.Item>
                )}
            />
        </Card>
    );
}

function ProjectCards({rows}: { rows: domain.ScoredRepository[] }) {
    if (rows.length === 0) {
        return null;
    }
    return (
        <Row gutter={[16, 16]}>
            {rows.map(({repository, score}) => (
                <Col xs={24} xl={12} key={repository.full_name}>
                    <Card
                        className="repo-card"
                        title={
                            <Space>
                                <GithubOutlined/>
                                <a href={repository.html_url} target="_blank" rel="noreferrer">{repository.full_name}</a>
                            </Space>
                        }
                    >
                        <Paragraph className="repo-description">{repository.description || 'No description provided.'}</Paragraph>
                        <Space wrap>
                            <Tag color="blue">{repository.language || 'Unknown'}</Tag>
                            <Tag>{repository.stars?.toLocaleString()} stars</Tag>
                            <Tag>{repository.forks?.toLocaleString()} forks</Tag>
                            <Tag color={levelColor(score.influence_level)}>Level {score.influence_level}</Tag>
                            <Tag color="purple">{score.difficulty}</Tag>
                        </Space>
                        <Divider/>
                        <Row gutter={16} align="middle">
                            <Col span={8}>
                                <Progress type="circle" percent={score.total_score} size={88}/>
                            </Col>
                            <Col span={16}>
                                <Text strong>{score.beginner_friendliness}</Text>
                                <Paragraph className="reason">{score.recommendation_reason}</Paragraph>
                            </Col>
                        </Row>
                    </Card>
                </Col>
            ))}
        </Row>
    );
}

function ScoreTable({rows}: { rows: domain.ScoredRepository[] }) {
    if (rows.length === 0) {
        return null;
    }
    return (
        <Card title="Score Breakdown" className="panel">
            <Table
                rowKey={(row) => row.repository.full_name}
                dataSource={rows}
                pagination={false}
                scroll={{x: 980}}
                columns={[
                    {
                        title: 'Repository',
                        dataIndex: ['repository', 'full_name'],
                        fixed: 'left',
                        render: (_value, row) => <a href={row.repository.html_url} target="_blank" rel="noreferrer">{row.repository.full_name}</a>,
                    },
                    {title: 'Total', dataIndex: ['score', 'total_score'], sorter: (a, b) => a.score.total_score - b.score.total_score},
                    {title: 'Activity', dataIndex: ['score', 'activity_score']},
                    {title: 'Popularity', dataIndex: ['score', 'popularity_score']},
                    {title: 'Learning', dataIndex: ['score', 'learning_value_score']},
                    {title: 'Contribution', dataIndex: ['score', 'contribution_friendliness_score']},
                    {title: 'Role', dataIndex: ['score', 'role_relevance_score']},
                    {title: 'Influence', dataIndex: ['score', 'influence_level']},
                    {title: 'Friendly', dataIndex: ['score', 'beginner_friendliness']},
                ]}
            />
        </Card>
    );
}

function ReportPreview({markdown}: { markdown: string }) {
    if (!markdown) {
        return (
            <Card className="empty-state">
                <Paragraph>Run discovery to generate a Markdown recommendation report.</Paragraph>
            </Card>
        );
    }
    return (
        <Card
            title="Markdown Preview"
            className="panel report-panel"
            extra={<Button icon={<ExportOutlined/>} onClick={() => downloadMarkdown(markdown)}>Export .md</Button>}
        >
            <ReactMarkdown>{markdown}</ReactMarkdown>
        </Card>
    );
}

function downloadMarkdown(markdown: string) {
    const blob = new Blob([markdown], {type: 'text/markdown;charset=utf-8'});
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'github-project-recommendation.md';
    link.click();
    URL.revokeObjectURL(url);
}

function levelColor(level: string) {
    switch (level) {
        case 'S':
            return 'red';
        case 'A':
            return 'orange';
        case 'B':
            return 'green';
        case 'C':
            return 'cyan';
        default:
            return 'default';
    }
}

export default App;
