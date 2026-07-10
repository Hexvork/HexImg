#pragma once

#include <QAbstractListModel>
#include <QVector>

struct ImageItem {
    QString path;
    QString fileName;
    QString outputPath;
};

class QueueModel final : public QAbstractListModel
{
    Q_OBJECT

public:
    enum Roles {
        PathRole = Qt::UserRole + 1,
        FileNameRole,
        OutputPathRole,
    };

    explicit QueueModel(QObject *parent = nullptr);

    int rowCount(const QModelIndex &parent = QModelIndex()) const override;
    QVariant data(const QModelIndex &index, int role = Qt::DisplayRole) const override;
    QHash<int, QByteArray> roleNames() const override;

    void setItems(const QVector<ImageItem> &items);
    const QVector<ImageItem> &items() const;

private:
    QVector<ImageItem> m_items;
};
