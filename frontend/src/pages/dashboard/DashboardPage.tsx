import { Typography } from "antd";

const { Title, Paragraph } = Typography;

function DashboardPage() {
  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>UBotHub Dashboard</Title>
      <Paragraph>
        Open platform for connecting chat bots with virtual avatars.
      </Paragraph>
    </div>
  );
}

export default DashboardPage;
