#include "queuemodel.h"

QueueModel::QueueModel(QObject *parent)
    : QAbstractListModel(parent)
{
}

int QueueModel::rowCount(const QModelIndex &parent) const
{
    return parent.isValid() ? 0 : m_items.size();
}

QVariant QueueModel::data(const QModelIndex &index, int role) const
{
    if (!index.isValid() || index.row() < 0 || index.row() >= m_items.size()) {
        return {};
    }

    const ImageItem &item = m_items.at(index.row());
    switch (role) {
    case PathRole:
        return item.path;
    case FileNameRole:
        return item.fileName;
    case OutputPathRole:
        return item.outputPath;
    default:
        return {};
    }
}

QHash<int, QByteArray> QueueModel::roleNames() const
{
    return {
        {PathRole, "path"},
        {FileNameRole, "fileName"},
        {OutputPathRole, "outputPath"},
    };
}

void QueueModel::setItems(const QVector<ImageItem> &items)
{
    beginResetModel();
    m_items = items;
    endResetModel();
}

const QVector<ImageItem> &QueueModel::items() const
{
    return m_items;
}
